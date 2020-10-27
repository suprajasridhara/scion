package plncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pln/internal/plncrypto"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
	"github.com/scionproto/scion/go/pln/internal/sqlite"
	"github.com/scionproto/scion/go/proto"
)

type PLNListHandler struct {
}

func (p PLNListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PLNListHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)
	message := r.FullMessage.(*ctrl.SignedPld)
	//verify source PLN AS signature on the message
	e := plncrypto.PLNEngine{Msgr: plnmsgr.Msgr, IA: plnmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	err := verifier.Verify(context.Background(), message.Blob, message.Sign)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
		log.Error("Certificate verification failed!", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	//verify signatures for every plnListEntry before inserting into db
	pld := &ctrl.Pld{}
	err = proto.ParseFromRaw(pld, message.Blob)
	if err != nil {
		log.Error("Parse from raw failed to parse to Plnlist", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	plnList := pld.Pln.PlnList.L

	for _, plnListEntry := range plnList {
		signedPld := &ctrl.SignedPld{}
		err = proto.ParseFromRaw(signedPld, plnListEntry.Raw)
		if err != nil {
			log.Error("Error parsing pcn entry", err)
			return nil
		}
		print(signedPld.Blob.Len())
		pldFromRaw := &ctrl.Pld{}
		err = proto.ParseFromRaw(pldFromRaw, signedPld.Blob)
		if err != nil {
			log.Error("Error parsing pcn entry", err)
			return nil
		}
		var pcnIAInt addr.IAInt
		pcnIAInt = addr.IAInt(pldFromRaw.Pcn.AddPLNEntryRequest.Entry.IA)
		pcnIA := pcnIAInt.IA()

		//verify pcnIA signature
		verifier := trust.Verifier{BoundIA: pcnIA, Engine: e}
		err := verifier.Verify(context.Background(), signedPld.Blob, signedPld.Sign)

		if err == nil {
			//persist entry in database
			err = insertIfNotExists(pldFromRaw.Pcn.AddPLNEntryRequest.Entry.PCNId,
				uint64(pcnIAInt), plnListEntry.Raw)
			if err != nil {
				log.Error("Error while inserting new entry")
				sendAck(proto.Ack_ErrCode_reject, err.Error())
			}
		} else {
			log.Error("Error verifying plnEntry signature", err)
		}

	}
	print(pld.Pln.PlnList.L[0].PCNId)
	return nil
}

func insertIfNotExists(pcnId string, plnIA uint64, raw []byte) error {
	//TODO (supraja): add if plnEntry not exists logic
	_, err := sqlite.Db.InsertNewPlnEntry(context.Background(), pcnId, plnIA, raw)
	return err
}
