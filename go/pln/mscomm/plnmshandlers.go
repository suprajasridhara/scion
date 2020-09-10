package mscomm

import (
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
	"github.com/scionproto/scion/go/pln/internal/sqlite"
	"github.com/scionproto/scion/go/proto"
)

type PlnListHandler struct {
}

func (a PlnListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PlnListHandler.Handle")
	ctx := r.Context()
	//logger := log.FromCtx(ctx)
	plnList, err := sqlite.Db.GetPlnList(ctx)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}
	var l []pln_mgmt.PlnListEntry
	for _, entry := range plnList {
		l = append(l, *pln_mgmt.NewPlnListEntry(uint8(entry.Id), uint64(entry.I), uint64(entry.A)))
	}

	plnL := pln_mgmt.NewPlnList(l)

	pld, err := pln_mgmt.NewPld(1, plnL)
	err = plnmsgr.Msgr.SendPlnList(ctx, pld, r.Peer, r.ID)
	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	return nil

}
