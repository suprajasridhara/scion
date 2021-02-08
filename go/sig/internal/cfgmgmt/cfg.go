package cfgmgmt

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sigjson"
	"github.com/scionproto/scion/go/sig/internal/mscomm"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
	"github.com/scionproto/scion/go/sig/internal/sigcrypto"
	"github.com/scionproto/scion/go/sig/internal/sqlite"
)

var (
	prefixFile string
)

//Init intiialises the package and also starts a poll on the file in prefixFilePath to push
//prefixes in prefixPushInterval intervals
func Init(ctx context.Context, cfgDir string, prefixFilePath string,
	prefixPushInterval time.Duration) error {

	sigcrypto.CfgDir = cfgDir
	sigSigner := &sigcrypto.SIGSigner{}
	sigSigner.Init(ctx, sigcmn.Msgr, sigcmn.IA, sigcrypto.CfgDir)
	sigcmn.CoreASes = sigSigner.SignedTRC.TRC.CoreASes

	prefixFile = prefixFilePath
	go func(ctx context.Context, prefixPushInterval time.Duration) {
		defer log.HandlePanic()
		pushPrefixes(ctx, prefixPushInterval)
	}(ctx, prefixPushInterval)

	return nil
}

func LoadCfg(cfg *sigjson.Cfg) error {
	log.Info("LodCfg: entering")
	asList := sigcmn.CoreASes
	success := false

	if len(asList) == 0 {
		return errors.New("No CoreASes found")
	}

	for _, as := range asList {
		ia := addr.IA{
			I: sigcmn.IA.I,
			A: as,
		}

		fm, err := mscomm.GetFullMap(ia)
		if err != nil {
			log.Error("Error getting map from ", "IA", ia, "err", err)
			continue // Try next core AS
		} else {
			success = true
		}

		if fm.Fm == nil {
			continue
		}

		for _, f := range fm.Fm {
			log.Info(f.Ia, f.Ip, f.Id)
			ia, err := addr.IAFromString(f.Ia)
			if err != nil {
				return common.NewBasicError("Unable to get IA from string", err, "raw", f.Ia)
			}
			ip, IPNet, err := net.ParseCIDR(f.Ip)
			if err != nil {
				return common.NewBasicError("Unable to parse IPnet string", err, "raw", f.Ip)
			}
			if !ip.Equal(IPNet.IP) {
				return common.NewBasicError(
					"Network is not canonical (should not be host address).",
					nil, "raw", f.Ip)
			}

			i := sigjson.IPNet(*IPNet)
			s := make([]*sigjson.IPNet, 1)
			s[0] = &i
			//if IA already existed in the old cfg, it is rewritten with the new prefix
			//fetched from MS
			cfg.ASes[ia] = &sigjson.ASEntry{Nets: s}
		}
	}

	if success {
		return nil
	} else {
		return common.NewBasicError("Unable to fetch map from core ASes ", nil)
	}

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
		if err := mscomm.AddASMap(ctx, p); err != nil {
			log.Error("Error pushing prefix "+p, "Error: ", err)
		}
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
