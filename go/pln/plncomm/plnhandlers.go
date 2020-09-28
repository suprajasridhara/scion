package plncomm

import (
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

type PLNListHandler struct {
}

func (p PLNListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: PLNListHandler.Handle")
	//verify signature on the message

	//verify signatures for every plnListEntry before inserting into db

	return nil
}
