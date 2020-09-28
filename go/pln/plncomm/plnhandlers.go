package plncomm

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

type PLNListHandler struct {
}

func (p PLNListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PLNListHandler.Handle")
	return nil
}
