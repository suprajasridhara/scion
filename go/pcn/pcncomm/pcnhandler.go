package pcncomm

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

type NodeListHandler struct {
}

func (n NodeListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: NodeListHandler.Handle")
	return nil
}
