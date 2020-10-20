package pcncomm

import (
	"context"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
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
	logger := log.FromCtx(ctx)
	asEntries, err := sqlite3.Db.GetNewEntries(context.Background()) //signed ASMapEntries in the form of SignedPld
	if err != nil {
		logger.Error("could not get entries from DB", "Err: ", err)
	}

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		logger.Error("error getting signer", "Err: ", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PushMSListRequest})

	entries := []ms_mgmt.SignedAsEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedAsEntry(asEntry.Blob, asEntry.Sign)
		entries = append(entries, *entry)
	}

	timestamp := time.Now()

	// pcnIA := pcns[0].PCNIA
	pcn := getRandomPCN(ctx)
	address := &snet.SVCAddr{IA: pcn.PCNIA, SVC: addr.SvcPCN}
	req := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), pcn.PCNId, entries, msmsgr.IA.String())
	print("Timestamp in MSLIST : ", timestamp.Unix())
	pld, err := ms_mgmt.NewPld(1, req)

	//TODO_Q (supraja): generate random id?
	reply, err := msmsgr.Msgr.SendSignedMSList(ctx, pld, address, 123)

	if err != nil {
		logger.Error("error getting reply from PCN", "Err: ", err)
	}

	//Validate PCN signature
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: address.IA, Engine: e}
	err = verifier.Verify(ctx, reply.Blob, reply.Sign)

	if err != nil {
		logger.Error("error verifying sign for PCN rep", "Err: ", err)
	}

	packed, err := proto.PackRoot(reply)
	_, err = sqlite3.Db.InsertPCNRep(context.Background(), packed)

	if err != nil {
		logger.Error("error persisting PCN rep", "Err: ", err)
	}

}

func PullNodeListEntry(ctx context.Context, query string) {
	logger := log.FromCtx(ctx)
	mscrypt := &mscrypto.MSSigner{}
	err := mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	if err != nil {
		logger.Error("error mscrypt init", "Err: ", err)
	}
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		logger.Error("error getting signer", "Err: ", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.NodeListEntryRequest})
	req := pcn_mgmt.NewNodeListEntryRequest(query)
	pld, err := pcn_mgmt.NewPld(1, req)
	if err != nil {
		logger.Error("error constructing payload", "Err: ", err)
	}
	pcn := getRandomPCN(ctx)
	address := &snet.SVCAddr{IA: pcn.PCNIA, SVC: addr.SvcPCN}

	nodeList, err := msmsgr.Msgr.SendNodeListRequest(ctx, pld, address, 123)
	print(nodeList.Sign)

}

func getRandomPCN(ctx context.Context) plncomm.PCN {
	logger := log.FromCtx(ctx)
	pcns, err := plncomm.GetPlnList(context.Background())
	if err != nil {
		logger.Error("error getting pcns", "Err: ", err)
	}
	//pick a random pcn to send signed list to
	randomIndex := rand.Intn(len(pcns))
	randomIndex = 1
	return pcns[randomIndex]
}
