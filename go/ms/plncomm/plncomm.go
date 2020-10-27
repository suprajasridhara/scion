package plncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pkg/trust/compat"
)

var PLNAddr addr.IA

type PCN struct {
	PCNId string
	PCNIA addr.IA
}

func GetPlnList(ctx context.Context) ([]PCN, error) {
	address := &snet.SVCAddr{IA: PLNAddr, SVC: addr.SvcPLN}

	plnListReq := pln_mgmt.NewPlnListReq("request")
	pld, err := pln_mgmt.NewPld(1, plnListReq)
	if err != nil {
		return nil, serrors.WrapStr("Error creating pln_mgmt pld", err)
	}

	signedPld, err := msmsgr.Msgr.GetPlnList(ctx, pld, address, 1)
	if err != nil {
		return nil, serrors.WrapStr("Error getting plnList from messenger", err)
	}
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: PLNAddr, Engine: e}

	verifiedPayload, err := signedPld.GetVerifiedPld(context.Background(),
		compat.Verifier{Verifier: verifier})

	if err != nil {
		return nil, serrors.WrapStr("Error getting verified payload", err)
	}
	plnList := verifiedPayload.Pln.PlnList

	pcns := []PCN{}

	for _, plnListEntry := range plnList.L {
		pcn := PCN{PCNId: plnListEntry.PCNId, PCNIA: addr.IAInt(plnListEntry.IA).IA()}
		pcns = append(pcns, pcn)
	}
	//Signature from PLN is validated, the list is now authenticated.

	return pcns, nil
}
