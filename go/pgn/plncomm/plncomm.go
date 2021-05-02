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
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pkg/trust/compat"
)

type PGN struct {
	//PGNId is the id of the PGN In the IA it is deployed
	PGNId string
	//PGNIA ia of the PGN
	PGNIA addr.IA
}

//AddPGNEntry registers the PGN with the PLN it was started with
func AddPGNEntry(ctx context.Context, pgnID string, ia addr.IA, plnIA addr.IA) error {
	addr := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	entry := pln_mgmt.NewPlnListEntry(pgnID, uint64(ia.IAInt()), nil)
	req := pgn_mgmt.NewAddPLNEntryRequest(*entry)

	pgncrypt := &pgncrypto.PGNSigner{}
	err := pgncrypt.Init(ctx, pgnmsgr.Msgr, pgnmsgr.IA, pgncrypto.CfgDir)
	if err != nil {
		log.Error("error getting pgncrypto", "err", err)
		return err
	}
	signer, err := pgncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer", "err", err)
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

/*GetPLNList The PGN sends the request using the infra.Messenger instance in msmsgr
package and verifies the origin of the response before processing it. It then returns the processed
list of PGN Id and IA objects to the calling function
*/
func GetPLNList(ctx context.Context, plnIA addr.IA) ([]PGN, error) {
	address := &snet.SVCAddr{IA: plnIA, SVC: addr.SvcPLN}

	plnListReq := pln_mgmt.NewPlnListReq("request")
	pld, err := pln_mgmt.NewPld(1, plnListReq)
	if err != nil {
		return nil, serrors.WrapStr("Error creating pln_mgmt pld", err)
	}

	signedPld, err := pgnmsgr.Msgr.GetPLNList(ctx, pld, address, rand.Uint64())
	if err != nil {
		return nil, serrors.WrapStr("Error getting plnList from messenger", err)
	}
	e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
	verifier := trust.Verifier{BoundIA: plnIA, Engine: e}

	verifiedPayload, err := signedPld.GetVerifiedPld(context.Background(),
		compat.Verifier{Verifier: verifier})

	if err != nil {
		return nil, serrors.WrapStr("Error getting verified payload", err)
	}
	plnList := verifiedPayload.Pln.PlnList

	pgns := []PGN{}
	for _, plnListEntry := range plnList.L {
		pgn := PGN{PGNId: plnListEntry.PGNId, PGNIA: addr.IAInt(plnListEntry.IA).IA()}
		pgns = append(pgns, pgn)
	}
	//Signature from PLN is validated, the list is now authenticated.

	return pgns, nil
}
