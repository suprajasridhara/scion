package mscmn

import (
	"github.com/scionproto/scion/go/lib/infra"
)

// type FullMapReqHandler struct {
// }

// func (f FullMapReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
// 	log.Info("Entering: FullMapReqHandler.Handle")
// 	fullMap := r.Message.(*ms_mgmt.Pld).FullMapReq.FullMap
// 	print(fullMap.Addr)
// 	return nil
// }

type IdIdHandler struct {
}

func (f IdIdHandler) Handle(r *infra.Request) *infra.HandlerResult {
	//log.Info("Entering: IdIdHandler.Handle")
	return nil
}
