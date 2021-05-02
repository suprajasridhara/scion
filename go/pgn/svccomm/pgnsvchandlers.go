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

package svccomm

import (
	"context"
	"encoding/csv"
	"os"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnentryhelper"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

const (
	Default_commitID = "TODO_commitID"
)

type AddPGNEntryReqHandler struct {
}

func (a AddPGNEntryReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	start := time.Now()
	log.Info("Entering: AddPGNEntryReqHandler.Handle")
	ctx := r.Context()

	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	pgnEntry := r.Message.(*pgn_mgmt.AddPGNEntryRequest)

	if err := pgnentryhelper.ValidatePGNEntry(pgnEntry,
		r.FullMessage.(*ctrl.SignedPld), true); err != nil {
		log.Error("Invalid PGNEntry", "Err: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//persist the entry
	e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
	pgnEntry.CommitID = generateCommitID()
	signedBlob, err := proto.PackRoot(r.FullMessage.(*ctrl.SignedPld))
	if err != nil {
		log.Error("Error packing signedPld ", "Err: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	err = pgnentryhelper.PersistEntry(pgnEntry, e, signedBlob)
	if err != nil {
		log.Error("Error persisting Entry ", "Err: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pgnRep := pgn_mgmt.NewPGNRep(*pgnEntry, uint64(time.Now().Unix()))
	pld, err := pgn_mgmt.NewPld(1, pgnRep)
	if err != nil {
		log.Error("Error getting pcn_mgmt.pld")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	signer, err := registerSigner(infra.PGNRep)
	if err != nil {
		log.Error("Error registering signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = *signer
	}

	rw.SendPGNRep(ctx, pld, infra.PGNRep)
	duration := time.Since(start)
	log.Info("Time elapsed AddPGNEntryReqHandler", "duration ", duration.String())

	f, err := os.OpenFile("times.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Cannot open times.csv", "Err ", err)
		return nil
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"AddPGNEntryReqHandler", time.Now().String(), duration.String()})
	if err := w.Error(); err != nil {
		log.Error("error writing csv:", "Error :", err)
	}
	return nil
}

type PGNEntryRequestHandler struct {
}

func (p PGNEntryRequestHandler) Handle(r *infra.Request) *infra.HandlerResult {
	start := time.Now()
	log.Info("Entering: PGNEntryRequestHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	message := r.FullMessage.(*ctrl.SignedPld)
	if err := pgnentryhelper.VerifyASSignature(ctx, message, requester.IA); err != nil {
		log.Error("Certificate verification failed!", "Error: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pgnEntryRequest := r.Message.(*pgn_mgmt.PGNEntryRequest)
	if pgnEntryRequest.EntryType == "" {
		pgnEntryRequest.EntryType = "%"
	}
	if pgnEntryRequest.SrcIA == "" {
		pgnEntryRequest.SrcIA = "%"
	}

	dbEntries, err := sqlite.Db.GetEntriesByTypeAndSrcIA(context.Background(), "%", "%")
	var l []common.RawBytes
	for _, dbEntry := range dbEntries {
		l = append(l, *dbEntry.SignedBlob)
	}

	signer, err := registerSigner(infra.PGNList)
	if err != nil {
		log.Error("Error registering signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	var emptyObjects []common.RawBytes
	if pgnEntryRequest.SrcIA == "%" {
		isds := pgnentryhelper.GetISDsInEntries(dbEntries)
		emptyObjects = pgnentryhelper.GetEmptyObjects(isds, signer)
	}
	pgnList := pgn_mgmt.NewPGNList(l, emptyObjects, uint64(time.Now().Unix()))
	pld, err := pgn_mgmt.NewPld(1, pgnList)
	if err != nil {
		log.Error("Error forming pgn_mgmt Pld", "Error: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = *signer
	}

	rw.SendPGNRep(ctx, pld, infra.PGNList)
	duration := time.Since(start)
	log.Info("Time elapsed MSPGNEntryRequestHandler", "duration ", duration.String())

	f, err := os.OpenFile("times.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Cannot open times.csv", "Err ", err)
		return nil
	}
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"MSPGNEntryRequestHandler", time.Now().String(), duration.String()})
	if err := w.Error(); err != nil {
		log.Error("error writing csv:", "Error :", err)
	}

	return nil
}

func generateCommitID() string {
	//For now this returns a static string. To be implemented when support for multiple commits
	//is added
	return Default_commitID
}

func verifyASSignature(ctx context.Context, message *ctrl.SignedPld, IA addr.IA) error {
	//Verify AS signature
	e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
	verifier := trust.Verifier{BoundIA: IA, Engine: e}
	return verifier.Verify(ctx, message.Blob, message.Sign)
}

func registerSigner(msgType infra.MessageType) (*trust.Signer, error) {
	pgncrypt := &pgncrypto.PGNSigner{}
	pgncrypt.Init(context.Background(), pgnmsgr.Msgr, pgnmsgr.IA, pgncrypto.CfgDir)
	signer, err := pgncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer", err)
	}
	pgnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{msgType})
	return &signer, nil
}
