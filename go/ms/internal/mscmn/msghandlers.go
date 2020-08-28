package mscmn

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

type FullMapReqHandler struct {
}

func (f FullMapReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: FullMapReqHandler.Handle")
	return nil
}
