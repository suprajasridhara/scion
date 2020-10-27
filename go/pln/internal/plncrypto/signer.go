package plncrypto

import (
	"context"
	"crypto"
	"crypto/x509"
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

type PLNSigner struct {
	Msgr      infra.Messenger
	IA        addr.IA
	SignerGen trust.SignerGenNoDB
	signedTRC cppki.SignedTRC
}

func (m *PLNSigner) Init(ctx context.Context, Msgr infra.Messenger,
	IA addr.IA, cfgDir string) error {
	m.Msgr = Msgr
	m.IA = IA
	t, err := m.getTRC()
	if err != nil {
		return serrors.WrapStr("Error in init plncrypto", err)
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
		return serrors.WrapStr("Error in init plncrypto", err)
	}
	for _, key := range keys {
		c[key], err = m.getChains(ctx, key)
		if err != nil {
			return serrors.WrapStr("Error in init plncrypto", err)
		}
	}
	m.SignerGen.PrivateKeys = keys
	m.SignerGen.Chains = c
	return nil
}

func (m *PLNSigner) getChains(ctx context.Context,
	key crypto.Signer) ([][]*x509.Certificate, error) {
	date := time.Now()
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	skid, _ := cppki.SubjectKeyID(key.Public())

	req := &cert_mgmt.ChainReq{RawIA: m.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	//TODO_Q (supraja): generate id randomly?
	rawChains, err := m.Msgr.GetCertChain(context.Background(), req, addr, 1234)
	if err != nil {
		return nil, serrors.WrapStr("Error in getChains", err)
	}

	return rawChains.Chains()

}
func (m *PLNSigner) getTRC() (cppki.SignedTRC, error) {
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	//TODO_Q (supraja): generate id randomly?
	//TODO (supraja): replace hard coded Ids
	encTRC, err := m.Msgr.GetTRC(context.Background(),
		&cert_mgmt.TRCReq{ISD: m.IA.I, Base: 1, Serial: 1}, addr, 1)
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch Core as", err)
	}
	return trc, nil
}
