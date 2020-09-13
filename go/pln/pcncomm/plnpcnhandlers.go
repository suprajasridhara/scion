package pcncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
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
	//logger := log.FromCtx(ctx)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	requester := r.Peer.(*snet.UDPAddr)
	m := r.FullMessage.(*ctrl.SignedPld)
	e := plncrypto.PLNEngine{Msgr: plnmsgr.Msgr, IA: plnmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	err := verifier.Verify(context.Background(), m.Blob, m.Sign) //TODO_Q (supraja): here, the validation is for the source AS because the signatures are with AS keys. Is it possible to get per entity keys with cpki?
	if err != nil {
		log.Error("Certificate verification failed!")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	plnEntry := r.Message.(*pcn_mgmt.AddPLNEntryRequest).Entry

	//validate that the IA in plnEntry is valid and equal to the requester IA
	if plnEntry.IA != uint64(requester.IA.IAInt()) {
		//IA in the entry is not the same as the requester IA, reject
		log.Error("IA in the entry is not the same as the requester IA")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	_, err = sqlite.Db.InsertNewPlnEntry(ctx, plnEntry.IA)
	if err != nil {
		log.Error("Error while inserting new entry")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	//TODO_Q (supraja): should this send a response?

	return nil
}
