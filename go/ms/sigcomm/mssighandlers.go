package sigcomm

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite"
	"github.com/scionproto/scion/go/ms/internal/validator"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pkg/trust/compat"
	"github.com/scionproto/scion/go/proto"
)

const (
	/* //TODO (supraja): this should be configurable and related
	to other timestamp values to be added later. See doc*/
	max_ms_as_add_time = 10 //time is in minutes.
)

//ASActionHandler is a handler for messages from SIG
type ASActionHandler struct {
}

//Handle handles an AS Action request from a SIG
//The SIG sends ms_mgmt.Pld with ms_mgmt.ASMapEntry. This handler verifies signatures,
//the ownership of the prefixes using RPKI and adds it to the
//new_entries database for further processing. It also sends back a ms_mgmt.MSRepToken
//to the SIG
func (a ASActionHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: ASActionHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)
	rw, _ := infra.ResponseWriterFromContext(r.Context())
	sendAck := messenger.SendAckHelper(ctx, rw)

	m := r.FullMessage.(*ctrl.SignedPld)
	if _, err := verify(requester.IA, m); err != nil {
		log.Error("Certificate verification failed!")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//Source IA validated here, make sure the source AS and the ASMap have the same address
	asMapEntry := r.Message.(*ms_mgmt.ASMapEntry)
	if requester.IA.String() != asMapEntry.Ia {
		log.Error("Invalid AS in map")
		sendAck(proto.Ack_ErrCode_reject, "Invalid AS in map")
		return nil
	}
	//All good. Source IA is the IA in the asMap as well.

	//Now validate the AS-IP mapping using an rpkivalidator
	//Do RPKI validation with a shell script for now
	//For now, the validator should take 2 arguments,
	//ASN and prefix and return "valid" if the mapping is valid
	cmd := exec.Command("/bin/sh", validator.Path, requester.IA.A.String(), asMapEntry.Ip[0])

	op, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err.Error())
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	if strings.TrimSpace(string(op)) != validator.EntryValid {
		log.Error("Not valid mapping")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//RPKI validation passed. Add entry to database to be read later
	//Full message from SIG. Contains *ms_mgmt.ASMapEntry along with signature
	packed, err := proto.PackRoot(m)

	x := &ctrl.SignedPld{}
	proto.ParseFromRaw(x, packed)

	if err != nil {
		log.Error("Unable to pack")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	_, err = sqlite.Db.InsertNewEntry(context.Background(), packed)
	if err != nil {
		log.Error("Error while inserting new entry")
		sendAck(proto.Ack_ErrCode_reject, err.Error())

	}

	//Done inserting token. Now send back token with message and signature
	signer, err := registerSigner(infra.ASActionReply)
	if err != nil {
		log.Error("Error registering signer", "err", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	timestamp := time.Now().Add(time.Minute * time.Duration(max_ms_as_add_time))
	rep := ms_mgmt.NewMSRepToken(packed, uint64(timestamp.Unix()))
	pld, err := ms_mgmt.NewPld(1, rep)

	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = *signer
	}

	rw.SendMSRep(context.Background(), pld, infra.ASActionReply)
	return nil
}

type FullMapReqHandler struct {
}

func (f FullMapReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: FullMapReqHandler.Handle")
	ctx := r.Context()
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	requester := r.Peer.(*snet.UDPAddr)

	m := r.FullMessage.(*ctrl.SignedPld)
	if _, err := verify(requester.IA, m); err != nil {
		log.Error("Certificate verification failed!")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	fullMapRes, err := sqlite.Db.GetFullMap(ctx)

	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	//Register signer with messenger and response writer
	signer, err := registerSigner(infra.MSFullMapReply)
	if err != nil {
		log.Error("Error registering signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = *signer
	}

	var fs []ms_mgmt.FullMap
	for _, fm := range fullMapRes {
		fs = append(fs, *ms_mgmt.NewFullMap(uint8(fm.ID), fm.IP.String, fm.IA.String))
	}

	fmRep := ms_mgmt.NewFullMapRep(fs)
	pld, err := ms_mgmt.NewPld(1, fmRep)
	err = rw.SendMSRep(context.Background(), pld, infra.MSFullMapReply)
	if err != nil {
		log.Error("Error sending fullMap", "err", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	return nil
}

func verify(ia addr.IA, spld *ctrl.SignedPld) (*ctrl.Pld, error) {
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: ia, Engine: e}
	pld, err := compat.Verifier{Verifier: verifier}.VerifyPld(context.Background(), spld)
	if err != nil {
		return nil, serrors.WrapStr("Invalid signature", err)
	}
	return pld, nil
}

func registerSigner(msgType infra.MessageType) (*trust.Signer, error) {
	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(context.Background(), msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{msgType})
	return &signer, nil
}
