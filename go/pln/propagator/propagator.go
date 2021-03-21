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
	if len(propTo) == 0 {
		log.Error(`No Core ASes to propagate list to. This might be because the configured Hops 
		is too small`)
	} else {
		for _, p := range propTo {
			address := &snet.SVCAddr{IA: p.IA(), SVC: addr.SvcPLN}
			err := plnmsgr.SendPLNList(address, rand.Uint64())
			if err != nil {
				log.Error("error sending list to "+address.String(), "error:", err)
			}
		}
	}
	return nil

}

/*asToPropTo returns the ASes to propagate PLN lists to. For this it takes
core segments and if the segment is shorter than p.N (core AS that is less
than p.N hops away) adds it to a slice to return. If there are no core ASes
less than p.N hops away it returns an empty slice and should be handled by
the caller.
p.N is configured on PLN startup. See config.PLNConf.Hops*/
func (p *Propagator) asToPropTo(recs []*seg.Meta) []addr.IAInt {
	var ias []addr.IAInt
	for _, rec := range recs {
		asEntries := rec.Segment.ASEntries
		if len(asEntries) <= int(p.N) { //the core AS is less than p.N hops away
			//check if the IA exists in the slice ias already
			newIA := asEntries[0].RawIA
			exists := false
			for _, ia := range ias {
				if ia == newIA {
					exists = true
					break
				}
			}
			if !exists {
				//ia is not in ias. Add it
				ias = append(ias, newIA)
			}
		}
	}
	return ias
}
