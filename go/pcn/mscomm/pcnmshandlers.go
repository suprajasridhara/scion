package mscomm

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

type MSListHandler struct {
}

func (m MSListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PlnListHandler.Handle")
	return nil
}
