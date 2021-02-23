package propagator

import (
	"context"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/seg"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
)

type Propagator struct {
	//N is the number of hops to propagate PLN entries to
	N uint16
}

//Start starts the ticker to propagate PLNLists in intervals specified by
//interval
func (p *Propagator) Start(ctx context.Context, interval time.Duration) {
	p.Run()
	propTicker := time.NewTicker(interval)
	for {
		select {
		case <-propTicker.C:
			err := p.Run()
			if err != nil {
				log.Error("Error in run ", "error:", err)
			}
		}
	}
}

//Run runs the Propagator. It fetches ASes that are less than equal to N hops away
func (p *Propagator) Run() error {
	msg := &path_mgmt.SegReq{RawSrcIA: plnmsgr.IA.IAInt(), RawDstIA: addr.IA{I: 0, A: 0}.IAInt()}
	csaddr := &snet.SVCAddr{IA: plnmsgr.IA, SVC: addr.SvcCS}

	rep, err := plnmsgr.Msgr.GetSegs(context.Background(), msg, csaddr, rand.Uint64())
	if err != nil {
		return serrors.New("Error getting segs", "error:", err)
	}

	recs := rep.Recs.Recs
	propTo := p.asToPropTo(recs)
	for _, p := range propTo {
		address := &snet.SVCAddr{IA: p.IA(), SVC: addr.SvcPLN}
		err := plnmsgr.SendPLNList(address, rand.Uint64())
		if err != nil {
			log.Error("error sending list to "+address.String(), "error:", err)
		}
	}
	return nil

}

func (p *Propagator) asToPropTo(recs []*seg.Meta) []addr.IAInt {
	var ias []addr.IAInt
	for _, rec := range recs {
		asEntries := rec.Segment.ASEntries
		if len(asEntries) <= int(p.N) {
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
