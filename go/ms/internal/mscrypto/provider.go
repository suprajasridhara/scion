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
// limitations under the License.package config
package mscrypto

import (
	"context"
	"crypto/x509"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pkg/trust"
)

//MSEngine implements the trust.Provider interface
type MSEngine struct {
	Msgr infra.Messenger
	IA   addr.IA
}

func (m MSEngine) NotifyTRC(ctx context.Context, trcId cppki.TRCID,
	o ...trust.Option) error {

	return nil
}

func (m MSEngine) GetChains(ctx context.Context, cq trust.ChainQuery,
	o ...trust.Option) ([][]*x509.Certificate, error) {

	date := time.Now()
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	skid := cq.SubjectKeyID
	req := &cert_mgmt.ChainReq{RawIA: cq.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	rawChains, err := m.Msgr.GetCertChain(ctx, req, addr, rand.Uint64())
	if err != nil {
		return nil, serrors.WrapStr("Unable to fetch Chains", err)
	}
	return rawChains.Chains()
}

func (m MSEngine) GetSignedTRC(ctx context.Context, trcId cppki.TRCID,
	o ...trust.Option) (cppki.SignedTRC, error) {

	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	encTRC, err := m.Msgr.GetTRC(context.Background(),
		&cert_mgmt.TRCReq{ISD: trcId.ISD, Base: trcId.Base, Serial: trcId.Serial},
		addr, rand.Uint64())
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch TRC", err)
	}
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch SignedTRC", err)
	}
	return trc, nil
}
