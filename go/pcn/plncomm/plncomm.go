package plncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/internal/types"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pkg/trust/compat"
)

func AddPCNEntry(ctx context.Context, pcnId string, ia addr.IA, plnIA addr.IA) error {
	addr := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	entry := pln_mgmt.NewPlnListEntry(pcnId, uint64(ia.IAInt()), nil)
	req := pcn_mgmt.NewAddPLNEntryRequest(*entry)

	pcncrypt := &pcncrypto.PCNSigner{}
	err := pcncrypt.Init(ctx, pcnmsgr.Msgr, pcnmsgr.IA, pcncrypto.CfgDir)
	if err != nil {
		log.Error("error getting pcncrypto", err)
		return err
	}
	signer, err := pcncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer", err)
		return err
	}
	pcnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.AddPLNEntryRequest})

	pld, err := pcn_mgmt.NewPld(1, req)

	if err != nil {
		log.Error("Error forming pcn_mgmt payload", "Err: ", err)
	}
	//TODO_Q (supraja): random id?
	err = pcnmsgr.Msgr.SendPLNEntry(ctx, pld, addr, 1)
	if err != nil {
		log.Error("Error sending PLNEntry ", "Error: ", err)
	}
	return err

}

func GetPlnList(ctx context.Context, plnIA addr.IA) ([]types.PCN, error) {
	address := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	plnListReq := pln_mgmt.NewPlnListReq("request")
	pld, err := pln_mgmt.NewPld(1, plnListReq)
	if err != nil {
		return nil, serrors.WrapStr("Error creating pln_mgmt pld", err)
	}
	signedPld, err := pcnmsgr.Msgr.GetPlnList(ctx, pld, address, 1)
	if err != nil {
		return nil, serrors.WrapStr("Error getting plnlist", err)
	}
	e := pcncrypto.PCNEngine{Msgr: pcnmsgr.Msgr, IA: pcnmsgr.IA}
	verifier := trust.Verifier{BoundIA: plnIA, Engine: e}
	verifiedPayload, err := signedPld.GetVerifiedPld(context.Background(),
		compat.Verifier{Verifier: verifier})
	if err != nil {
		return nil, serrors.WrapStr("Error getting verifiedPayload for plnlist", err)

	}
	plnList := verifiedPayload.Pln.PlnList

	pcns := []types.PCN{}

	for _, plnListEntry := range plnList.L {
		pcn := types.PCN{PCNId: plnListEntry.PCNId, PCNIA: addr.IAInt(plnListEntry.IA).IA()}
		pcns = append(pcns, pcn)
	}
	//Signature from PLN is validated, the list is now authenticated.

	return pcns, nil
}
