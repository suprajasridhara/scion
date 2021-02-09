package pgncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
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

type AddPLNEntryHandler struct {
}

func (a AddPLNEntryHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: AddPLNEntryHandler.Handle")
	ctx := r.Context()
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	requester := r.Peer.(*snet.UDPAddr)
	m := r.FullMessage.(*ctrl.SignedPld)
	e := plncrypto.PLNEngine{Msgr: plnmsgr.Msgr, IA: plnmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	err := verifier.Verify(context.Background(), m.Blob, m.Sign)
	if err != nil {
		log.Error("Certificate verification failed!")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	plnEntry := r.Message.(*pgn_mgmt.AddPLNEntryRequest).Entry

	//validate that the IA in plnEntry is valid and equal to the requester IA
	if plnEntry.IA != uint64(requester.IA.IAInt()) {
		//IA in the entry is not the same as the requester IA, reject
		log.Error("IA in the entry is not the same as the requester IA")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	packed, err := proto.PackRoot(m)
	if err != nil {
		log.Error("Error while packing message")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	_, err = sqlite.Db.InsertNewPLNEntry(context.Background(), plnEntry.PGNId, plnEntry.IA, packed)

	if err != nil {
		log.Error("Error while inserting new entry", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	sendAck(proto.Ack_ErrCode_ok, "OK")

	log.Info("Exiting: AddPLNEntryHandler.Handle")

	return nil
}
