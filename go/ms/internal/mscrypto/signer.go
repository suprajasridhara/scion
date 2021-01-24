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
	"crypto"
	"crypto/x509"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pkg/trust"
)

//MSSigner is used by the Mapping Service to sign messages as well as verify signatures
type MSSigner struct {
	Msgr      infra.Messenger
	IA        addr.IA
	SignerGen trust.SignerGenNoDB
	signedTRC cppki.SignedTRC
}

//Init initializes MSSigner
func (m *MSSigner) Init(ctx context.Context, Msgr infra.Messenger,
	IA addr.IA, cfgDir string) error {
	m.Msgr = Msgr
	m.IA = IA
	t, err := m.getTRC()
	if err != nil {
		return serrors.WrapStr("Error in init MSSigner", err)
	}
	m.signedTRC = t
	s := make([]cppki.SignedTRC, 1)
	s[0] = m.signedTRC
	m.SignerGen = trust.SignerGenNoDB{
		IA: m.IA,
		KeyRing: LoadingRing{
			Dir: filepath.Join(cfgDir, "crypto/as"),
		},
		SignedTRCs: s,
	}
	c := make(map[crypto.Signer][][]*x509.Certificate)
	keys, err := m.SignerGen.KeyRing.PrivateKeys(ctx)
	if err != nil {
		return serrors.WrapStr("Error in init MSSigner", err)
	}
	for _, key := range keys {
		c[key], err = m.getChains(ctx, key)
		if err != nil {
			return serrors.WrapStr("Error in init MSSigner", err)
		}
	}
	m.SignerGen.PrivateKeys = keys
	m.SignerGen.Chains = c
	return nil
}

func (m *MSSigner) getChains(ctx context.Context,
	key crypto.Signer) ([][]*x509.Certificate, error) {
	
	date := time.Now()
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	skid, _ := cppki.SubjectKeyID(key.Public())

	req := &cert_mgmt.ChainReq{RawIA: m.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	rawChains, err := m.Msgr.GetCertChain(ctx, req, addr, rand.Uint64())
	if err != nil {
		return nil, serrors.WrapStr("Error in getChains", err)
	}

	return rawChains.Chains()

}
func (m *MSSigner) getTRC() (cppki.SignedTRC, error) {
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	encTRC, err := m.Msgr.GetTRC(context.Background(),
		&cert_mgmt.TRCReq{ISD: m.IA.I, Base: 1, Serial: 1}, addr, rand.Uint64())
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch TRC", err)
	}
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch Core as", err)
	}
	return trc, nil
}
