package pcncomm

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
	"github.com/scionproto/scion/go/ms/plncomm"
)

func SendSignedList(ctx context.Context, interval time.Duration) {
	pushSignedPrefix(ctx)
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			pushSignedPrefix(ctx)
		}
	}
}

func pushSignedPrefix(ctx context.Context) {
	asEntries, err := sqlite3.Db.GetNewEntries(context.Background()) //signed ASMapEntries in the form of SignedPld
	if err != nil {
		log.Error("could not get entries from DB")
	}

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer")
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PushMSListRequest})

	entries := []ms_mgmt.SignedAsEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedAsEntry(asEntry.Blob, asEntry.Sign)
		entries = append(entries, *entry)
	}

	timestamp := time.Now()

	pcns, err := plncomm.GetPlnList(ctx)
	if err != nil {
		log.Error("error getting pcns")
	}

	address := &snet.SVCAddr{IA: pcns[0].PCNIA, SVC: addr.SvcPCN}
	req := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), pcns[0].PCNId, entries)
	pld, err := ms_mgmt.NewPld(1, req)

	//TODO_Q (supraja): generate random id?
	msmsgr.Msgr.SendSignedMSList(ctx, pld, address, 123)

}
