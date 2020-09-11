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
)

var PLNAddr addr.IA

func GetPlnList(ctx context.Context) error {
	addr := &snet.SVCAddr{IA: PLNAddr, SVC: addr.SvcPLN}

	plnListReq := pln_mgmt.NewPlnListReq("request")
	pld, err := pln_mgmt.NewPld(1, plnListReq)
	if err != nil {
		return serrors.WrapStr("Error creating pln_mgmt pld", err)
	}
	signedPld, err := msmsgr.Msgr.GetPlnList(ctx, pld, addr, 1)
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: PLNAddr, Engine: e}
	err = verifier.Verify(ctx, signedPld.Blob, signedPld.Sign)
	if err != nil {
		return serrors.WrapStr("Invalid signature", err)
	}

	//Signature from PLN is validated, the list is now authenticated.
	return nil
}
