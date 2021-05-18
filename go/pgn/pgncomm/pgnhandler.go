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
	"encoding/csv"
	"os"
	"time"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnentryhelper"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

type PGNEntryHandler struct {
	SignedPld *ctrl.SignedPld
	Pld       *ctrl.Pld
}

func (n PGNEntryHandler) Handle(r *infra.Request) *infra.HandlerResult {
	start := time.Now()
	log.Info("Entering: PGNEntryHandler.Handle")
	ctx := context.Background()
	//requester := r.Peer.(*snet.UDPAddr)

	//Verify node list signature by pgn on the list
	//message := r.FullMessage.(*ctrl.SignedPld)
	message := n.SignedPld
	e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
	verifier := trust.Verifier{BoundIA: pgnmsgr.IA, Engine: e}
	err := verifier.Verify(ctx, message.Blob, message.Sign)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
		log.Error("Certificate verification failed!", "Error: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pldList := n.Pld
	// err = proto.ParseFromRaw(pldList, message.Blob)
	// if err != nil {
	// 	log.Error("Error decerializing control payload", "Error: ", err)
	// 	return nil
	// }
	c := make(chan int, 50)
	for i, l := range pldList.Pgn.PGNList.L {
		c <- i
		go func(ch chan int, l common.RawBytes) {
			defer log.HandlePanic()
			signedPGNEntry := &ctrl.SignedPld{}
			err = proto.ParseFromRaw(signedPGNEntry, l)
			if err != nil {
				log.Error("Error getting signedPGNEntry", "Error: ", err)
			}
			pld := &ctrl.Pld{}
			proto.ParseFromRaw(pld, signedPGNEntry.Blob)
			r := pld.Pgn.AddPGNEntryRequest
			if err = pgnentryhelper.ValidatePGNEntry(r, signedPGNEntry, false); err != nil {
				log.Error("Error validating signatures", "Error: ", err)
			}

			//all verification done. Persist PGNEntry
			e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
			if err = pgnentryhelper.PersistEntry(r, e, l); err != nil {
				log.Error("Error persisting PGNEntry ", "Error: ", err)
			}
			<-ch
		}(c, l)
	}

	//process empty objects
	for _, emptyObjectBlob := range pldList.Pgn.PGNList.EmptyObjects {
		signedEmptyObject := &ctrl.SignedPld{}
		err = proto.ParseFromRaw(signedEmptyObject, emptyObjectBlob)
		if err != nil {
			log.Error("Error getting signedPGNEntry", "Error: ", err)
			continue
		}
		pld := &ctrl.Pld{}
		proto.ParseFromRaw(pld, signedEmptyObject.Blob)
		sqlite.Db.InsertEmptyObject(context.Background(), emptyObjectBlob, pld.Pgn.EmptyObject.Isd, pld.Pgn.EmptyObject.Timestamp)
	}

	duration := time.Since(start)
	log.Info("Time elapsed 8-PGNEntryHandler", "duration ", duration.String())

	f, err := os.OpenFile("times.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Cannot open times.csv", "Err ", err)
		return nil
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"8-PGNEntryHandler", time.Now().String(), duration.String()})
	if err := w.Error(); err != nil {
		log.Error("error writing csv:", "Error :", err)
	}
	return nil
}
