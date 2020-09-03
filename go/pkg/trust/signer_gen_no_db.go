// Copyright 2020 Anapaya Systems
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

package trust

import (
	"context"
	"crypto"
	"crypto/x509"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pkg/trust/internal/metrics"
)

//TODO (supraja): SignerGen uses a database. Since non cs services might not have a trust database they can query the cs for trc and cert chains and generate signatures for messages. Is this the correct behavior or should we just use a trust db?
//This is a copy of SignerGen in signer_gen.go with DB replaced with SignedTRC and Chains
// SignerGenNoDB generates signers from the keys available in key dir. It does not require a DB connection instead takes in SignedTRC assumed to be active and Chains
type SignerGenNoDB struct {
	IA          addr.IA
	KeyRing     KeyRing
	PrivateKeys []crypto.Signer
	SignedTRCs  []cppki.SignedTRC
	Chains      map[crypto.Signer][][]*x509.Certificate
	// DB      DB // FIXME(roosd): Eventually this should use a crypto provider

}

// Generate fetches private keys from the key ring and searches active
// certificate chains that authenticate the corresponding public key. The
// returned signer uses the private key which is backed by the certificate chain
// with the highest expiration time.
func (s SignerGenNoDB) Generate(ctx context.Context) (Signer, error) {
	l := metrics.SignerLabels{}
	//keys, err := s.KeyRing.PrivateKeys(ctx)
	keys := s.PrivateKeys
	// if err != nil {
	// 	metrics.Signer.Generate(l.WithResult(metrics.ErrKey)).Inc()
	// 	return Signer{}, err
	// }
	if len(keys) == 0 {
		metrics.Signer.Generate(l.WithResult(metrics.ErrKey)).Inc()
		return Signer{}, serrors.New("no private key found")
	}

	// trcs, res, err := activeTRCs(ctx, s.DB, s.IA.I)
	// if err != nil {
	// 	metrics.Signer.Generate(l.WithResult(res)).Inc()
	// 	return Signer{}, serrors.WrapStr("loading TRC", err)
	// }

	// Search the private key that has a certificate that expires the latest.
	var best *Signer
	for _, key := range keys {
		signer, err := s.bestForKey(ctx, key)
		if err != nil {
			metrics.Signer.Generate(l.WithResult(metrics.ErrDB)).Inc()
			return Signer{}, err
		}
		if signer == nil {
			continue
		}
		if best != nil && signer.Expiration.Before(best.Expiration) {
			continue
		}
		best = signer
	}
	if best == nil {
		metrics.Signer.Generate(l.WithResult(metrics.ErrNotFound)).Inc()
		return Signer{}, serrors.New("no certificate found", "num_private_keys", len(keys))
	}
	metrics.Signer.Generate(l.WithResult(metrics.Success)).Inc()
	return *best, nil
}

func (s *SignerGenNoDB) bestForKey(ctx context.Context, key crypto.Signer) (*Signer, error) {
	// FIXME(roosd): We currently take the sha1 sum of the public key.
	// The final implementation needs to be smarter than that, but this
	// requires a proper design that also considers certificate renewal.
	// skid, err := cppki.SubjectKeyID(key.Public())
	// if err != nil {
	// 	return nil, nil
	// }
	// chains, err := s.DB.Chains(ctx, ChainQuery{
	// 	IA:           s.IA,
	// 	SubjectKeyID: skid,
	// 	Date:         time.Now(),
	// })
	// if err != nil {
	// 	// TODO	metrics.Signer.Generate(l.WithResult(metrics.ErrDB)).Inc()
	// 	return nil, err
	// }

	chains, _ := s.Chains[key]
	chain := bestChainNoDB(&s.SignedTRCs[0].TRC, chains)
	if chain == nil && len(s.SignedTRCs) == 1 {
		return nil, nil
	}
	var inGrace bool
	// Attempt to find a chain that is verifiable only in grace period. If we
	// have not found a chain yet.
	if chain == nil && len(s.SignedTRCs) == 2 {
		chain = bestChainNoDB(&s.SignedTRCs[1].TRC, chains)
		if chain == nil {
			return nil, nil
		}
		inGrace = true
	}
	id, expiry := s.SignedTRCs[0].TRC.ID, min(chain[0].NotAfter, s.SignedTRCs[0].TRC.Validity.NotAfter)
	if inGrace {
		id, expiry = s.SignedTRCs[1].TRC.ID, min(chain[0].NotAfter, s.SignedTRCs[0].TRC.GracePeriodEnd())
	}
	return &Signer{
		PrivateKey:   key,
		Hash:         crypto.SHA512,
		IA:           s.IA,
		TRCID:        id,
		SubjectKeyID: chain[0].SubjectKeyId,
		Expiration:   expiry,
		ChainValidity: cppki.Validity{
			NotBefore: chain[0].NotBefore,
			NotAfter:  chain[0].NotAfter,
		},
		InGrace: inGrace,
	}, nil
}

func bestChainNoDB(trc *cppki.TRC, chains [][]*x509.Certificate) []*x509.Certificate {
	opts := cppki.VerifyOptions{TRC: trc}
	var best []*x509.Certificate
	for _, chain := range chains {
		if err := cppki.VerifyChain(chain, opts); err != nil {
			continue
		}
		if len(best) > 0 && chain[0].NotAfter.Before(best[0].NotAfter) {
			continue
		}
		best = chain
	}
	return best
}
