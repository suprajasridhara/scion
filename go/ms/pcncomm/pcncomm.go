package pcncomm

import (
	"context"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
	"github.com/scionproto/scion/go/ms/plncomm"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
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
		//log.Error("could not get entries from DB", err)
	}

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		//log.Error("error getting signer", err)
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
		//log.Error("error getting pcns", err)
	}

	//pick a random pcn to send signed list to

	randomIndex := rand.Intn(len(pcns))
	pcnIA := pcns[randomIndex].PCNIA

	// pcnIA := pcns[0].PCNIA
	address := &snet.SVCAddr{IA: pcnIA, SVC: addr.SvcPCN}
	req := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), pcns[randomIndex].PCNId, entries, msmsgr.IA.String())
	print("Timestamp in MSLIST : ", timestamp.Unix())
	pld, err := ms_mgmt.NewPld(1, req)

	//TODO_Q (supraja): generate random id?
	reply, err := msmsgr.Msgr.SendSignedMSList(ctx, pld, address, 123)

	if err != nil {
		// log.Error("error getting reply from PCN", err)
	}

	//Validate PCN signature
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: pcnIA, Engine: e}
	// msmsgr.Msgr.UpdateVerifier(verifier)
	err = verifier.Verify(ctx, reply.Blob, reply.Sign)

	if err != nil {
		//log.Error("error verifying sign for PCN rep", err)
	}

	packed, err := proto.PackRoot(reply)
	_, err = sqlite3.Db.InsertPCNRep(context.Background(), packed)

	if err != nil {
		//log.Error("error persisting PCN rep", err)
	}

}
