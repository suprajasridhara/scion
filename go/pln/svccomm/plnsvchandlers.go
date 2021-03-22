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

	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pln/internal/plncrypto"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
	"github.com/scionproto/scion/go/proto"
)

//SvcListHandler is the Handler to handle PLNList requests from services
type SvcListHandler struct {
}

func (a SvcListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: SvcListHandler.Handle")
	ctx := r.Context()
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)

	signer, err := registerSigner(infra.PlnListReply)
	if err != nil {
		log.Error("Error registering signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	switch t := rw.(type) {
	case *messenger.QUICResponseWriter:
		t.Signer = *signer
	}

	pld, err := plnmsgr.GetPLNListAsPld(r.ID)
	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}

	err = rw.SendPLNList(context.Background(), pld)

	if err != nil {
		sendAck(proto.Ack_ErrCode_reject, err.Error())
	}
	return nil

}

func registerSigner(msgType infra.MessageType) (*trust.Signer, error) {
	plncrypt := &plncrypto.PLNSigner{}
	plncrypt.Init(context.Background(), plnmsgr.Msgr, plnmsgr.IA, plncrypto.CfgDir)
	signer, err := plncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer", err)
	}
	plnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{msgType})
	return &signer, nil
}
