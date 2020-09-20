package pcncrypto

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pkg/trust"
)

type PCNEngine struct {
	Msgr infra.Messenger
	IA   addr.IA
}

func (m PCNEngine) NotifyTRC(ctx context.Context, trcId cppki.TRCID,
	o ...trust.Option) error {
	//TODO (supraja): implement this
	return nil
}

func (m PCNEngine) GetChains(ctx context.Context, cq trust.ChainQuery,
	o ...trust.Option) ([][]*x509.Certificate, error) {
	date := time.Now()
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	skid := cq.SubjectKeyID
	req := &cert_mgmt.ChainReq{RawIA: cq.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	//TODO_Q (supraja): generate id randomly?
	rawChains, err := m.Msgr.GetCertChain(ctx, req, addr, 1234)
	if err != nil {
		return nil, serrors.WrapStr("Unable to fetch Chains", err)
	}
	return rawChains.Chains()
}

func (m PCNEngine) GetSignedTRC(ctx context.Context, trcId cppki.TRCID,
	o ...trust.Option) (cppki.SignedTRC, error) {
	addr := &snet.SVCAddr{IA: m.IA, SVC: addr.SvcCS}
	//TODO_Q (supraja): generate id randomly?
	encTRC, err := m.Msgr.GetTRC(context.Background(), &cert_mgmt.TRCReq{ISD: trcId.ISD, Base: trcId.Base, Serial: trcId.Serial}, addr, 1)
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch SignedTRC", err)
	}
	return trc, nil
}
