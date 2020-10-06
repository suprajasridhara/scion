package pcncomm

import (
	"context"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/plncomm"
)

//TODO (supraja): read this from config file
const (
	n = 1
)

func BroadcastNodeList(ctx context.Context, interval time.Duration, plnIA addr.IA) {
	err := sendSignedPCNList(ctx, plnIA)
	if err != nil {
		log.Error("error in broadcast node list", err)
	}
	pushTicker := time.NewTicker(interval * time.Second)
	for {
		select {
		case <-pushTicker.C:
			err = sendSignedPCNList(ctx, plnIA)
			if err != nil {
				log.Error("error in broadcast node list", err)
			}
		}
	}
}

func sendSignedPCNList(ctx context.Context, plnIA addr.IA) error {
	pcns, err := plncomm.GetPlnList(ctx, plnIA)
	if err != nil {
		return serrors.WrapStr("Error getting pln list", err)
	}

	if n > len(pcns) {
		return serrors.WrapStr("n is greater than nuymber of pcns in PLN list", err)
	}

	var randIs []int
	for i := 0; i < n; i++ {
		r := rand.Intn(len(pcns))
		if !contains(randIs, r) {
			randIs = append(randIs, r)
		} else {
			i--
		}
	}

	for i := range randIs {
		pcn := pcns[i]
		pcnmsgr.SendNodeList(context.Background(), pcn.PCNIA)
	}

	return nil

}

func contains(l []int, elem int) bool {
	for e := range l {
		if e == elem {
			return true
		}
	}
	return false
}
