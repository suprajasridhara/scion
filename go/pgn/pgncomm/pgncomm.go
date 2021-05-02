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

package pgncomm

import (
	"context"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
	"github.com/scionproto/scion/go/pgn/plncomm"
)

//N is the number of PGNs the list is propagated to in every time interval
var N int

func BroadcastNodeList(ctx context.Context, interval time.Duration, plnIA addr.IA) {
	log.Info("Entering: BroadcastNodeList")
	err := sendPGNList(ctx, plnIA)
	if err != nil {
		log.Error("error in broadcast node list", "err", err)
	}
	log.Error("BroadcastNodeList", "interval", interval.String())

	pushTicker := time.NewTicker(interval)
	for {
		select {
		case <-pushTicker.C:
			err = sendPGNList(ctx, plnIA)
			if err != nil {
				log.Error("error in broadcast node list", "err", err)
			}
		}
	}
}

func sendPGNList(ctx context.Context, plnIA addr.IA) error {
	log.Info("Entering: sendPGNList")
	dbEntries, err := sqlite.Db.GetEntriesByTypeAndSrcIA(context.Background(), "%", "%")
	if err != nil {
		return serrors.WrapStr("Error getting full node list", err)
	}

	if len(dbEntries) > 0 {
		pgns, err := plncomm.GetPLNList(ctx, plnIA)
		if err != nil {
			return serrors.WrapStr("Error getting pln list", err)
		}

		if N > len(pgns) {
			return serrors.WrapStr("n is greater than number of PGNs in PLN list", err)
		}

		var randIs []int
		for i := 0; i < N; i++ { //pick n pgns at random from pgns
			r := rand.Intn(len(pgns))
			if !contains(randIs, r) && !pgns[r].PGNIA.Equal(pgnmsgr.IA) {
				randIs = append(randIs, r)
			} else {
				i--
			}
		}

		var l []common.RawBytes
		for _, dbEntry := range dbEntries {
			l = append(l, *dbEntry.SignedBlob)
		}
		pgnList := pgn_mgmt.NewPGNList(l, uint64(time.Now().Unix()))
		pld, err := pgn_mgmt.NewPld(1, pgnList)
		if err != nil {
			return serrors.WrapStr("Error forming pgn_mgmt Pld", err)
		}
		for _, i := range randIs {
			pgn := pgns[i]
			address := &snet.SVCAddr{IA: pgn.PGNIA, SVC: addr.SvcPGN}

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
			pgnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PGNList})
			err = pgnmsgr.Msgr.SendPGNRep(context.Background(), pld, address,
				rand.Uint64(), infra.PGNList)
			if err != nil {
				log.Error("Error sending pgn list", "err", err)
			}
		}
	}
	log.Info("Exiting: sendPGNList")

	return nil

}

func contains(l []int, elem int) bool {
	for e := range l {
		if e == elem {
			return true
		}
	}
	return false
}
