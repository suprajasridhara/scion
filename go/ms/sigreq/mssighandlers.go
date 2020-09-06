package sigreq

import (
	"context"
	"fmt"
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
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
	"github.com/scionproto/scion/go/ms/internal/validator"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

const (
	max_ms_as_add_time = 10 //time is in minutes. TODO (supraja): this should be configurable and related to other timestamp values to be added later. See doc
)

type FullMapReqHandler struct {
}

func (f FullMapReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: FullMapReqHandler.Handle")
	ctx := r.Context()
	//logger := log.FromCtx(ctx)
	fullMapRes, err := sqlite3.Db.GetFullMap(ctx)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}
	var fs []ms_mgmt.FullMap
	for _, fm := range fullMapRes {
		//TODO (supraja): handle conversions properly here
		fs = append(fs, *ms_mgmt.NewFullMap(uint8(fm.Id), fm.IP.String, fm.IA.String))
	}

	fmrep := ms_mgmt.NewFullMapRep(fs)

	pld, err := ms_mgmt.NewPld(1, fmrep)
	err = msmsgr.Msgr.SendFullMap(ctx, pld, r.Peer, r.ID)
	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	return nil
}

type ASActionHandler struct {
}

func (a ASActionHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: ASActionHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)
	m := r.FullMessage.(*ctrl.SignedPld)
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	// msmsgr.Msgr.UpdateVerifier(verifier)
	err := verifier.Verify(ctx, m.Blob, m.Sign)
	rw, _ := infra.ResponseWriterFromContext(ctx)
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
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//Source IA is the IA in the asMap as well. Now validate the AS-IP mapping using an rpkivalidator
	//TODO_Q (supraja): Is this ok to Assume the AS is a BGP style AS?

	//Do RPKI validation with a shell script for now

	//TODO (supraja): read this correctly from config file.
	//For now, the validator should take 2 arguments, asn and prefix and return "valid" if the mapping is valid
	//TODO (supraja): find a better way to do this
	cmdStr := validator.Path + " " + requester.IA.A.String() + " " + asMapEntry.Ip[0]
	cmd := exec.Command("/bin/sh", "-c", cmdStr)

	if err != nil {
		log.Error(err.Error())
	}
	op, err := cmd.Output()
	if err != nil {
		log.Error(err.Error())
		x := err.Error()
		fmt.Println(x)
	}

	if strings.TrimSpace(string(op)) != validator.EntryValid {
		log.Error("Not valid mapping")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//RPKI validation passed. Add entry to database to be read later

	packed, err := proto.PackRoot(m)

	x := &ctrl.SignedPld{}
	proto.ParseFromRaw(x, packed)

	if err != nil {
		log.Error("Unable to pack")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	_, err = sqlite3.Db.InsertNewEntry(context.Background(), packed)
	if err != nil {
		log.Error("Error while inserting new entry")
		sendAck(proto.Ack_ErrCode_reject, err.Error())

	}

	// pld, err := sqlite3.Db.GetNewEntryById(context.Background(), 1)
	// print(pld)
	// if err != nil {
	// 	print(err.Error())
	// }
	// y := bytes.Equal(pld.Blob, m.Blob)
	// print(y)

	//Done inserting token. Now send back token with message and signature

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())

	}

	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.ASActionReply})

	timestamp := time.Now().Add(time.Minute * time.Duration(max_ms_as_add_time))
	//TODO (supraja): handle int64 to uint64 conversion correctly
	rep := ms_mgmt.NewMSRepToken(packed, uint64(timestamp.Unix()))
	pld, err := ms_mgmt.NewPld(1, rep)
	msmsgr.Msgr.SendASMSRepToken(context.Background(), pld, r.Peer, r.ID)
	return nil
}
