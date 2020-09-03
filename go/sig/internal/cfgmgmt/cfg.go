package cfgmgmt

import (
	"context"
	"crypto"
	"crypto/x509"
	"net"
	"path/filepath"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/cert_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scrypto/cppki"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/sigjson"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
)

const (
	ADD_AS_ENTRY = "add_as_entry"
)

var (
	SignerGen trust.SignerGenNoDB
	signedTRC cppki.SignedTRC
)

func Init(ctx context.Context, cfgDir string) {
	//TODO (supraja): handle error
	signedTRC, _ = getTRC()

	s := make([]cppki.SignedTRC, 1)
	s[0] = signedTRC
	SignerGen = trust.SignerGenNoDB{
		IA: sigcmn.IA,
		KeyRing: LoadingRing{
			Dir: filepath.Join(cfgDir, "crypto/as"),
		},
		SignedTRCs: s,
	}
	c := make(map[crypto.Signer][][]*x509.Certificate)
	keys, _ := SignerGen.KeyRing.PrivateKeys(ctx)
	for _, key := range keys {
		c[key], _ = getChains(ctx, key)
	}
	SignerGen.PrivateKeys = keys
	SignerGen.Chains = c

}

func getChains(ctx context.Context, key crypto.Signer) ([][]*x509.Certificate, error) {
	date := time.Now()
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	skid, _ := cppki.SubjectKeyID(key.Public())

	req := &cert_mgmt.ChainReq{RawIA: sigcmn.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	//TODO (supraja): fix id
	rawChains, err := sigcmn.Msgr.GetCertChain(ctx, req, addr, 1234)
	if err != nil {
		//TODO (supraja): handle error properly
		log.Error(err.Error())
	}

	return rawChains.Chains()

}
func getTRC() (cppki.SignedTRC, error) {
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	//TODO (supraja): read from config
	encTRC, err := sigcmn.Msgr.GetTRC(context.Background(), &cert_mgmt.TRCReq{ISD: 1, Base: 1, Serial: 1}, addr, 1)
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch Core as", err)
	}
	return trc, nil
}

func GetCoreASs() []addr.AS {
	// addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	// //TODO (supraja): read from config
	// encTRC, err := sigcmn.Msgr.GetTRC(context.Background(), &cert_mgmt.TRCReq{ISD: 1, Base: 1, Serial: 1}, addr, 1)
	// trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	// if err != nil {
	// 	return nil, serrors.WrapStr("Unable to fetch Core as", err)
	// }
	return signedTRC.TRC.CoreASes
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

func AddASMap(ctx context.Context, ip string) error {
	ia := addr.IA{
		I: sigcmn.IA.I,
		A: GetCoreASs()[0],
	}
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	//TODO (supraja): replace hardcoded Ids
	timestamp := uint64(time.Now().UnixNano())
	asEntry := ms_mgmt.NewASMapEntry([]string{ip}, sigcmn.IA.String(), timestamp, ADD_AS_ENTRY)
	signer, err := SignerGen.Generate(ctx)
	if err != nil {
		return serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	sigcmn.Msgr.UpdateSigner(signer, []infra.MessageType{infra.ASActionRequest})

	pld, err := ms_mgmt.NewPld(1, asEntry)
	sigcmn.Msgr.SendASAction(ctx, pld, addr, 1)
	return nil
}

func LoadCfg(cfg *sigjson.Cfg) error {
	log.Info("LodCfg: entering")
	asList := GetCoreASs()

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
