package sigreq

import (
	"context"

	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
)

type FullMapReqHandler struct {
}

func (f FullMapReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: FullMapReqHandler.Handle")
	//_ := r.Message.(*ms_mgmt.Pld).FullMapReq.FullMap
	fullMapRes, err := sqlite3.Db.GetFullMap(context.Background())
	if err != nil {
		//TODO (supraja): return error here
	}
	//print(fullMapRes[0].IP.String)
	var fs []ms_mgmt.FullMap
	for _, fm := range fullMapRes {
		//TODO (supraja): handle conversions properly here
		fs = append(fs, *ms_mgmt.NewFullMap(uint8(fm.Id), fm.IP.String, fm.IA.String))
	}

	fmrep := ms_mgmt.NewFullMapRep(fs)

	pld, err := ms_mgmt.NewPld(1, fmrep)
	err = msmsgr.Msgr.SendFullMap(context.Background(), pld, r.Peer, r.ID)
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}

type ASActionHandler struct {
}

func (a ASActionHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: ASActionHandler.Handle")

	return nil
}
