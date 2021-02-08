package cfgmgmt

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/sig/internal/mscomm"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
	"github.com/scionproto/scion/go/sig/internal/sigcrypto"
	"github.com/scionproto/scion/go/sig/internal/sqlite"
)

var (
	prefixFile string
)

//Init initialises the package and also starts a poll on the file in prefixFilePath to push
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
