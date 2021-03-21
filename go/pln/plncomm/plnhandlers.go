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

//PLNListHandler is the handler for PLNLists that are broadcast by PLN services
type PLNListHandler struct {
}

//Handle handlers PLNList that is sent by other PLNs. It verifies the signature of the
//source PLN and signatures on each of the entries before uppdating the PLN DB with any
//new entries received
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
			log.Error("Error parsing SignedPld", err)
			return nil
		}
		pldFromRaw := &ctrl.Pld{}
		err = proto.ParseFromRaw(pldFromRaw, signedPld.Blob)
		if err != nil {
			log.Error("Error parsing PGN entry", err)
			return nil
		}
		var pgnIAInt addr.IAInt
		pgnIAInt = addr.IAInt(pldFromRaw.Pgn.AddPLNEntryRequest.Entry.IA)
		pgnIA := pgnIAInt.IA()

		//verify pgnIA signature
		verifier := trust.Verifier{BoundIA: pgnIA, Engine: e}
		err := verifier.Verify(context.Background(), signedPld.Blob, signedPld.Sign)

		if err == nil {
			//persist entry in database
			_, err = sqlite.Db.InsertNewPLNEntry(context.Background(),
				pldFromRaw.Pgn.AddPLNEntryRequest.Entry.PGNId, uint64(pgnIAInt), plnListEntry.Raw)
			if err != nil {
				log.Error("Error while inserting new entry")
			}
		} else {
			log.Error("Error verifying plnEntry signature", "error: ", err)
		}

	}
	return nil
}