package cfgmgmt

import (
	"bufio"
	"context"
	"crypto"
	"crypto/x509"
	"math/rand"
	"net"
	"os"
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
	"github.com/scionproto/scion/go/proto"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
	"github.com/scionproto/scion/go/sig/internal/sigcrypto"
	"github.com/scionproto/scion/go/sig/internal/sqlite"
)

const (
	ADD_AS_ENTRY = "add_as_entry"
)

var (
	SignerGen  trust.SignerGenNoDB
	CoreASes   []addr.AS
	prefixFile string
)

func Init(ctx context.Context, cfgDir string, prefixFilePath string,
	prefixPushInterval time.Duration) error {
	signedTRC, err := getTRC()
	CoreASes = signedTRC.TRC.CoreASes
	if err != nil {
		return err
	}
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

	//push prefixes
	prefixFile = prefixFilePath
	go func(ctx context.Context, prefixPushInterval time.Duration) {
		defer log.HandlePanic()
		pushPrefixes(ctx, prefixPushInterval)
	}(ctx, prefixPushInterval)
	return nil

}
func pushPrefixes(ctx context.Context, interval time.Duration) {
	addPrefixes(ctx)
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			addPrefixes(ctx)
		}
	}

}

func addPrefixes(ctx context.Context) {
	pushed, _ := sqlite.Db.GetPushedPrefixes(ctx)
	read, _ := readPrefixes(prefixFile)
	newPrefixes := listDiff(read, pushed) //performs read - pushed
	for _, p := range newPrefixes {
		//TODO (supraja): handle error
		AddASMap(ctx, p)
	}
}

func listDiff(l1 []string, l2 []string) []string {
	res := []string{}
	for _, one := range l1 {
		exists := false
		for _, two := range l2 {
			if one == two {
				exists = true
			}
		}
		if !exists {
			res = append(res, one)
		}
	}
	return res
}
func readPrefixes(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getChains(ctx context.Context, key crypto.Signer) ([][]*x509.Certificate, error) {
	date := time.Now()
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	skid, _ := cppki.SubjectKeyID(key.Public())

	req := &cert_mgmt.ChainReq{RawIA: sigcmn.IA.IAInt(), SubjectKeyID: skid, RawDate: date.Unix()}
	//TODO_Q (supraja): random id?
	rawChains, err := sigcmn.Msgr.GetCertChain(ctx, req, addr, rand.Uint64())
	if err != nil {
		return nil, serrors.WrapStr("Error getting cert chains", err)
	}

	return rawChains.Chains()

}
func getTRC() (cppki.SignedTRC, error) {
	addr := &snet.SVCAddr{IA: sigcmn.IA, SVC: addr.SvcCS}
	//TODO_Q (supraja): Where do we get Base and Serial from?
	encTRC, err := sigcmn.Msgr.GetTRC(context.Background(),
		&cert_mgmt.TRCReq{ISD: sigcmn.IA.I, Base: 1, Serial: 1}, addr, 1)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch TRC", err)
	}
	trc, err := cppki.DecodeSignedTRC(encTRC.RawTRC)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("Unable to fetch SignedTRC", err)
	}
	return trc, nil
}

func GetFullMap(ia addr.IA) (*ms_mgmt.FullMapRep, error) {
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	//TODO_Q (supraja): random values for id?
	pld, err := ms_mgmt.NewPld(1, ms_mgmt.NewFullMapReq(1))
	if err != nil {
		return nil, serrors.WrapStr("Unable to create payload", err)
	}
	return sigcmn.Msgr.GetFullMap(context.Background(), pld, addr, 1)
}

func AddASMap(ctx context.Context, ip string) error {
	ia := addr.IA{
		I: sigcmn.IA.I,
		A: CoreASes[0],
	}
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	//TODO_Q (supraja): random ids?
	timestamp := uint64(time.Now().UnixNano())
	asEntry := ms_mgmt.NewASMapEntry([]string{ip}, sigcmn.IA.String(), timestamp, ADD_AS_ENTRY)
	signer, err := SignerGen.Generate(ctx)
	if err != nil {
		return serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	sigcmn.Msgr.UpdateSigner(signer, []infra.MessageType{infra.ASActionRequest})
	pld, err := ms_mgmt.NewPld(1, asEntry)
	if err != nil {
		return serrors.WrapStr("Error forming ms_mgmt payload", err)

	}
	rep, err := sigcmn.Msgr.SendASAction(ctx, pld, addr, 1)
	if err != nil {
		return serrors.WrapStr("Error sending AS Action", err)

	}
	//verify signatures here
	//TODO_Q (supraja): Should we use the messenger verifier.
	//It is currently tightly coupled with control plane messages that have a trust store

	e := sigcrypto.SIGEngine{Msgr: sigcmn.Msgr, IA: sigcmn.IA}
	verifier := trust.Verifier{BoundIA: ia, Engine: e}
	err = verifier.Verify(context.Background(), rep.Blob, rep.Sign)

	if err != nil {
		return serrors.WrapStr("Invalid signature", err)
	}

	//The signature is validated. store the MSToken for future use
	packed, err := proto.PackRoot(rep)
	if err != nil {
		return serrors.WrapStr("Error paking reply to insert into db", err)
	}
	_, err = sqlite.Db.InsertNewMSToken(context.Background(), packed)
	if err != nil {
		return serrors.WrapStr("Error storing MS token into db", err)
	}

	//add the pushed prefix into the database
	err = insertIntoDB(ip)
	if err != nil {
		return serrors.WrapStr("Error storing pushed prefix into db", err)
	}

	return nil
}

func insertIntoDB(prefix string) error {
	_, err := sqlite.Db.InsertNewPushedPrefix(context.Background(), prefix)
	return err
}
func LoadCfg(cfg *sigjson.Cfg) error {
	log.Info("LodCfg: entering")
	asList := CoreASes
	success := false
	//TODO (supraja): impelemnt wait mechanism after timeout from each core AS.
	//For now contact one Core AS assuming the TRC had atleast one Core AS

	/*//TODO_Q (supraja): on debugging I see that an scmp packet is returned when
	the service doesnt exist, but the dispatcher has no handler, so the error is
	just logged and the dispatcher returns false and continues. Messenger waits for reply
	on channels that are never returned to.
	error logged: scmp packet received, but no handler found scmp.Hdr="Class=ROUTING(1)
	Type=UNREACH_HOST(1) TotalLen=64B Checksum=89bb Timestamp=2020-09-10 07:55:33.827096 +0000 UTC"
	src="{1-ff00:0:112 fd00:f00d:cafe::7f00:9}*/

	for _, as := range asList {
		ia := addr.IA{
			I: sigcmn.IA.I,
			A: as,
		}

		//ia, _ := addr.IAFromString("1-ff00:0:112")

		fm, err := GetFullMap(ia)

		if err != nil {
			continue //TRY next core AS
		} else {
			success = true
		}

		//TODO (supraja): based on file as whitelist or blacklist handle original cfg values here
		for _, f := range fm.Fm {
			ia, err := addr.IAFromString(f.Ia)
			if err != nil {
				return common.NewBasicError("Unable to get IA from string", err, "raw", f.Ia)
			}
			ip, ipnet, err := net.ParseCIDR(f.Ip)
			if err != nil {
				return common.NewBasicError("Unable to parse IPnet string", err, "raw", f.Ip)
			}
			if !ip.Equal(ipnet.IP) {
				return common.NewBasicError(
					"Network is not canonical (should not be host address).",
					nil, "raw", f.Ip)
			}
			//TODO (supraja): if IA exists, add logic to handle
			i := sigjson.IPNet(*ipnet)
			s := make([]*sigjson.IPNet, 1)
			s[0] = &i
			cfg.ASes[ia] = &sigjson.ASEntry{Nets: s}
		}
		// }
	}

	if success {
		return nil
	} else {
		return common.NewBasicError("Unable to fetch map from core ASs ", nil)
	}

}
