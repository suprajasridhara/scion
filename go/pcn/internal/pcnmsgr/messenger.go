package pcnmsgr

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/sqlite"
)

var Msgr infra.Messenger
var IA addr.IA
var Id string

func SendNodeList(ctx context.Context, pcnIA addr.IA) error {
	fullNodeList, err := sqlite.Db.GetFullNodeList(context.Background())
	if err != nil {
		return serrors.WrapStr("Error getting full node list", err)
	}

	var nodeListEntries []pcn_mgmt.NodeListEntry
	for _, nle := range fullNodeList {
		nodeListEntries = append(nodeListEntries, *pcn_mgmt.NewNodeListEntry(nle.MsList, nle.CommitId.String))
	}

	timestamp := time.Now()
	nodeList := pcn_mgmt.NewNodeList(nodeListEntries, uint64(timestamp.Unix()))

	pld, err := pcn_mgmt.NewPld(1, nodeList)

	pcncrypt := &pcncrypto.PCNSigner{}
	pcncrypt.Init(ctx, Msgr, IA, pcncrypto.CfgDir)
	signer, err := pcncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		//log.Error("error getting signer", err)
	}
	Msgr.UpdateSigner(signer, []infra.MessageType{infra.NodeList})

	address := &snet.SVCAddr{IA: pcnIA, SVC: addr.SvcPCN}

	//TODO_Q (supraja): random Id?
	err = Msgr.SendNodeList(context.Background(), pld, address, 1234323)
	if err != nil {
		return serrors.WrapStr("Error sending node list", err)
	}
	return nil
}
