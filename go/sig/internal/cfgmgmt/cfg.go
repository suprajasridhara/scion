package cfgmgmt

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/sigjson"
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

func GetFullMap(ia addr.IA) (*ms_mgmt.FullMapRep, error) {
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	//TODO (supraja): replace hardcoded Ids
	pld, err := ms_mgmt.NewPld(1, ms_mgmt.NewFullMapReq(1))
	if err != nil {
		return nil, serrors.WrapStr("Unable to create payload", err)
	}
	return sigcmn.Msgr.GetFullMap(context.Background(), pld, addr, 1)
}

func LoadCfg(cfg *sigjson.Cfg) error {
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

	fm, err := GetFullMap(ia)

	if err != nil {
		print(err.Error())
		return serrors.WrapStr("Unable to get full map", err)
	}

	//TODO (supraja): based on file as whitelist or blacklist handle original cfg values here
	for _, f := range fm.Fm {
		//TODO (supraja): handle error
		ia, _ := addr.IAFromString(f.Ia)
		ip, ipnet, err := net.ParseCIDR(f.Ip)
		if err != nil {
			return common.NewBasicError("Unable to parse IPnet string", err, "raw", f.Ip)
		}
		if !ip.Equal(ipnet.IP) {
			return common.NewBasicError("Network is not canonical (should not be host address).",
				nil, "raw", f.Ip)
		}
		//TODO (supraja): if IA exists, add logic to handle
		i := sigjson.IPNet(*ipnet)
		s := make([]*sigjson.IPNet, 1)
		s[0] = &i
		cfg.ASes[ia] = &sigjson.ASEntry{Nets: s}
	}
	return nil
}
