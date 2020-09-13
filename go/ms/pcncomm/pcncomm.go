package pcncomm

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
)

func SendSignedList(ctx context.Context, interval time.Duration) {
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			pushSignedPrefix(ctx)
		}
	}
}

func pushSignedPrefix(ctx context.Context) {
	_, err := sqlite3.Db.GetNewEntries(context.Background()) //signed ASMapEntries in the form of SignedPld
	if err != nil {
		log.Error("could not get entries from DB")
	}

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer")
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PushMSListRequest})

}
