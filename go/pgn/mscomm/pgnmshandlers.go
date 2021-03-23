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

package mscomm

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
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
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

const (
	Default_commitID = "TODO_commitID"
)

type AddPGNEntryReqHandler struct {
}

func (a AddPGNEntryReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: AddPGNEntryReqHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)

	err := verifyASSignature(ctx, r.FullMessage.(*ctrl.SignedPld), requester.IA)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
		log.Error("Certificate verification failed!", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pgnEntry := r.Message.(*pgn_mgmt.AddPGNEntryRequest)

	valid, err := pgnentryhelper.IsValidPGNEntry(pgnEntry)
	if err != nil {
		log.Error("Invalid PGNEntry", "Err: ", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	if valid {
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
