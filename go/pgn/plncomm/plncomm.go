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
package plncomm

import (
	"context"
	"math/rand"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
)

//AddPGNEntry registers the PGN with the PLN it was started with
func AddPGNEntry(ctx context.Context, pgnID string, ia addr.IA, plnIA addr.IA) error {
	addr := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	entry := pln_mgmt.NewPlnListEntry(pgnID, uint64(ia.IAInt()), nil)
	req := pgn_mgmt.NewAddPLNEntryRequest(*entry)

	pgncrypt := &pgncrypto.PGNSigner{}
	err := pgncrypt.Init(ctx, pgnmsgr.Msgr, pgnmsgr.IA, pgncrypto.CfgDir)
	if err != nil {
		log.Error("error getting pgncrypto", err)
		return err
	}
	signer, err := pgncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer", err)
		return err
	}
	pgnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.AddPLNEntryRequest})

	pld, err := pgn_mgmt.NewPld(1, req)
	if err != nil {
		log.Error("Error forming pgn_mgmt payload", "Err: ", err)
	}

	err = pgnmsgr.Msgr.SendPLNEntry(ctx, pld, addr, rand.Uint64())
	if err != nil {
		log.Error("Error sending PLNEntry ", "Error: ", err)
	}
	return err
}
