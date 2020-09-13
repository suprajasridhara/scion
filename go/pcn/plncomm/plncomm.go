package plncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
)

func AddPCNEntry(ctx context.Context, pcnId string, ia addr.IA, plnIA addr.IA) error {
	addr := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	entry := pln_mgmt.NewPlnListEntry(pcnId, uint64(ia.IAInt()))
	req := pcn_mgmt.NewAddPLNEntryRequest(*entry)

	pcncrypt := &pcncrypto.PCNSigner{}
	err := pcncrypt.Init(ctx, pcnmsgr.Msgr, pcnmsgr.IA, pcncrypto.CfgDir)
	if err != nil {
		log.Error("error getting pcncrypto")
		return err
	}
	signer, err := pcncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer")
		return err
	}
	pcnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.AddPLNEntryRequest})

	pld, err := pcn_mgmt.NewPld(1, req)

	//TODO_Q (supraja): random id?
	err = pcnmsgr.Msgr.SendPLNEntry(ctx, pld, addr, 1)
	if err != nil {
		log.Error("error", err)
	}
	return err

}
