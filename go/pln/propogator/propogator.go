package propogator

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/seg"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
)

//TODO_Q (supraja): read this from config?
const n = 3

type Propogator struct {
}

func (p *Propogator) Run() {
	msg := &path_mgmt.SegReq{RawSrcIA: plnmsgr.IA.IAInt(), RawDstIA: addr.IA{I: 0, A: 0}.IAInt()}
	csaddr := &snet.SVCAddr{IA: plnmsgr.IA, SVC: addr.SvcCS}

	rep, err := plnmsgr.Msgr.GetSegs(context.Background(), msg, csaddr, 1234343)
	if err != nil {
		log.Error("error", err)
	}

	//print(rep.Req.RawDstIA)

	recs := rep.Recs.Recs
	propTo := asToPropTo(recs)
	for _, p := range propTo {
		// print(p.IA().A.String())
		// print("\n")
		address := &snet.SVCAddr{IA: p.IA(), SVC: addr.SvcPLN}
		//TODO_Q (supraja): random id?
		err := plnmsgr.SendPLNList(address, 123423)
		if err != nil {
			log.Error("error sending list to "+address.String(), err)
		}
	}

}

func asToPropTo(recs []*seg.Meta) []addr.IAInt {
	var ias []addr.IAInt
	for _, rec := range recs {
		asEntries := rec.Segment.ASEntries
		if len(asEntries) <= n {
			newIA := asEntries[0].RawIA
			exists := false
			for _, ia := range ias {
				if ia == newIA {
					exists = true
					break
				}
			}
			if !exists {
				ias = append(ias, newIA)
			}
		}

	}
	return ias
}
