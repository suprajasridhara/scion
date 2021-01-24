package sigcomm

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite"
	"github.com/scionproto/scion/go/ms/internal/validator"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

const (
	/* //TODO (supraja): this should be configurable and related
	to other timestamp values to be added later. See doc*/
	max_ms_as_add_time = 10 //time is in minutes.
)

type ASActionHandler struct {
}

func (a ASActionHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: ASActionHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)
	m := r.FullMessage.(*ctrl.SignedPld)
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	err := verifier.Verify(context.Background(), m.Blob, m.Sign)
	rw, _ := infra.ResponseWriterFromContext(r.Context())
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
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
	cmdStr := validator.Path + " " + requester.IA.A.String() + " " + asMapEntry.Ip[0]
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	op, err := cmd.Output()
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
	mscrypt := &mscrypto.MSSigner{}
	err = mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	if err != nil {
		log.Error("error initializing crypto", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.ASActionReply})
	timestamp := time.Now().Add(time.Minute * time.Duration(max_ms_as_add_time))
	rep := ms_mgmt.NewMSRepToken(packed, uint64(timestamp.Unix()))
	pld, err := ms_mgmt.NewPld(1, rep)

	//rw.SendFullMap(context.Background(), pld)
	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = signer
	}

	rw.SendMSRep(context.Background(), pld, infra.ASActionReply)
	return nil
}
