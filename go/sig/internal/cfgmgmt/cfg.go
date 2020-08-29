package cfgmgmt

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
)

func GetCoreASs() ([]addr.AS, error) {
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	//TODO (supraja): read from config
	encTRC, err := sigcmn.Msgr.GetTRC(context.Background(), &cert_mgmt.TRCReq{ISD: 1, Base: 1, Serial: 1}, addr, 1)
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return nil, serrors.WrapStr("Unable to fetch Core as", err)
	}
	return trc.TRC.CoreASes, nil
}

func GetFullMap(ia addr.IA) error {
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	//TODO (supraja): read from config
	err := sigcmn.Msgr.GetFullMap(context.Background(), ms_mgmt.NewFullMapReq(1), addr, 1)
	if err != nil {
		return serrors.WrapStr("Unable to fetch TRC", err)
	}
	return nil
}

func LoadCfg() error {
	log.Info("LodCfg: entering")
	asList, err := GetCoreASs()
	if err != nil {
		return serrors.WrapStr("Unable to Load config", err)
	}
	//TODO (supraja): impelemnt wait mechanism after timeout from each core AS. For now contact one Core AS assuming the TRC had atleast one Core AS
	ia := addr.IA{
		I: sigcmn.IA.I,
		A: asList[0],
	}

	GetFullMap(ia)
	return nil
}
