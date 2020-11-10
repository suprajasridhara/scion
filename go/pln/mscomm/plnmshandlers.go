package mscomm

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
	"github.com/scionproto/scion/go/proto"
)

type PlnListHandler struct {
}

func (a PlnListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PlnListHandler.Handle")
	ctx := r.Context()
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	err := plnmsgr.SendPLNList(r.Peer, r.ID)
	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	return nil

}
