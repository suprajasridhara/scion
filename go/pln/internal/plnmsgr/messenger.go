// Copyright 2021 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func GetPLNListAsPld(id uint64) (*pln_mgmt.Pld, error) {
	var pld *pln_mgmt.Pld
	plnList, err := sqlite.Db.GetPLNList(context.Background())
	if err != nil {
		return nil, err
	}
	var l []pln_mgmt.PlnListEntry
	added := make(map[string]bool)
	for _, entry := range plnList {
		if !added[entry.PgnID] {
			l = append(l, *pln_mgmt.NewPlnListEntry(entry.PgnID, uint64(entry.IA), entry.Raw))
			added[entry.PgnID] = true
		}
	}

	if len(l) > 0 {
		plnL := pln_mgmt.NewPlnList(l)

		plncrypt := &plncrypto.PLNSigner{}
		plncrypt.Init(context.Background(), Msgr, IA, plncrypto.CfgDir)
		signer, err := plncrypt.SignerGen.Generate(context.Background())
		if err != nil {
			log.Error("Error getting signer", "error: ", err)
			return nil, err
		}

		plncrypt.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PlnListReply})

		pld, err = pln_mgmt.NewPld(1, plnL)
		if err != nil {
			return nil, err
		}
	}
	return pld, nil
}

//SendPLNList sends PLNList to addr
func SendPLNList(addr net.Addr, id uint64) error {
	pld, err := GetPLNListAsPld(id)
	if err != nil {
		return err
	}
	if pld != nil {
		err = Msgr.SendPLNList(context.Background(), pld, addr, id)
		if err != nil {
			return err
		}
	}
	return nil
}
