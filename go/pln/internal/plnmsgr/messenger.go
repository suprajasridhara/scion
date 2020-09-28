package plnmsgr

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/pln/internal/plncrypto"
	"github.com/scionproto/scion/go/pln/internal/sqlite"
)

var Msgr infra.Messenger
var IA addr.IA

func SendPLNList(addr net.Addr, id uint64) error {
	plnList, err := sqlite.Db.GetPlnList(context.Background())
	if err != nil {
		return err
	}
	var l []pln_mgmt.PlnListEntry
	added := make(map[string]bool)
	for _, entry := range plnList {
		if !added[entry.PcnId] {
			l = append(l, *pln_mgmt.NewPlnListEntry(entry.PcnId, uint64(entry.IA)))
			added[entry.PcnId] = true
		}
	}

	if len(l) > 0 {
		plnL := pln_mgmt.NewPlnList(l)

		plncrypt := &plncrypto.PLNSigner{}
		plncrypt.Init(context.Background(), Msgr, IA, plncrypto.CfgDir)
		signer, err := plncrypt.SignerGen.Generate(context.Background())
		if err != nil {
			log.Error("error getting signer", err)
			return err
			//sendAck(proto.Ack_ErrCode_reject, err.Error())

		}

		plncrypt.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PlnListReply})

		pld, err := pln_mgmt.NewPld(1, plnL)
		err = Msgr.SendPlnList(context.Background(), pld, addr, id)
		if err != nil {
			return err
			//sendAck(proto.Ack_ErrCode_reject, err.Error())
		}
	}
	return nil

}
