package cfgmgmt

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
)

func GetCoreASs() ([]addr.AS, error) {
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	//TODO (supraja): read from config
	encTRC, err := sigcmn.Msgr.GetTRC(context.Background(), &cert_mgmt.TRCReq{ISD: 2, Base: 1, Serial: 1}, addr, 1)
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return nil, serrors.WrapStr("Unable to fetch TRC", err)
	}
	return trc.TRC.CoreASes, nil
}

func LoadCfg() {
	asList := GetCoreASs()

}
