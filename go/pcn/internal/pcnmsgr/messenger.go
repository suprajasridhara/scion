package pcnmsgr

import (
	"context"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/sqlite"
)

var Msgr infra.Messenger
var IA addr.IA
var Id string

func SendNodeList(ctx context.Context, address net.Addr,
	fullNodeList []sqlite.NodeListEntry, id uint64) error {
	if len(fullNodeList) > 0 {
		var nodeListEntries []pcn_mgmt.NodeListEntry
		for _, nle := range fullNodeList {
			nodeListEntries = append(nodeListEntries,
				*pcn_mgmt.NewNodeListEntry(
					common.RawBytes(*nle.MsList), nle.CommitId.String))
		}

		timestamp := time.Now()
		nodeList := pcn_mgmt.NewNodeList(nodeListEntries, uint64(timestamp.Unix()))

		pld, err := pcn_mgmt.NewPld(1, nodeList)

		pcncrypt := &pcncrypto.PCNSigner{}
		pcncrypt.Init(ctx, Msgr, IA, pcncrypto.CfgDir)
		signer, err := pcncrypt.SignerGen.Generate(context.Background())
		if err != nil {
			return serrors.WrapStr("Error getting signer", err)
		}
		Msgr.UpdateSigner(signer, []infra.MessageType{infra.NodeList})
		err = Msgr.SendNodeList(context.Background(), pld, address, id)
		if err != nil {
			return serrors.WrapStr("Error sending node list", err)
		}
	}
	return nil
}
