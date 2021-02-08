package mscomm

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
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

//AddASMap creates a ms_mgmt.Pld with ms_mgmt.ASMapEntry and sends
//it to a MS in sigcmn.CoreASes. It then reads ms_mgmt.MSRepToken
//from the MS and stores it in ms_token table
func AddASMap(ctx context.Context, ip string) error {
	success := false
	var rep *ctrl.SignedPld
	var ia addr.IA
	for _, as := range sigcmn.CoreASes {
		ia = addr.IA{
			I: sigcmn.IA.I,
			A: as,
		}
		addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
		timestamp := uint64(time.Now().UnixNano())
		asEntry := ms_mgmt.NewASMapEntry([]string{ip}, sigcmn.IA.String(), timestamp, ADD_AS_ENTRY)

		if err := registerSigner(infra.ASActionRequest); err != nil {
			return err
		}

		pld, err := ms_mgmt.NewPld(1, asEntry)
		if err != nil {
			return serrors.WrapStr("Error forming ms_mgmt payload", err)

		}

		rep, err = sigcmn.Msgr.SendASAction(ctx, pld, addr, rand.Uint64())
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				//The messenger was not able to establish a connection with a MS in the IA.
				//This could be because there is no MS registered with the dispatcher in the IA
				//Or the MS is down. Try again with a different core AS
				log.Error("Not able to connect to MS", "IA ", ia.String())
			} else {
				//If the error is something other than not being able to reach a MS. Break and report
				//that the sending failed. This will be retried in the next time interval
				return serrors.WrapStr("Error sending AS Action", err)
			}
		} else {
			success = true
			break
		}
	}
	if success { //The prefix was successfully pushed to the MS and a token was received
		return doSuccess(rep, ia, ip)
	}

	return serrors.WrapStr("Pushing to MS was unsuccessfull",
		errors.New(`Could not connect to any MS in the core ASes. This could be because of 
		very small configured connect_period or no MS instance deployed in any core AS`))
}

func doSuccess(rep *ctrl.SignedPld, ia addr.IA, ip string) error {
	e := sigcrypto.SIGEngine{Msgr: sigcmn.Msgr, IA: sigcmn.IA}
	verifier := trust.Verifier{BoundIA: ia, Engine: e}
	err := verifier.Verify(context.Background(), rep.Blob, rep.Sign)
	if err != nil {
		return serrors.WrapStr("Invalid signature", err)
	}

	//The signature is validated. Store the MSToken for future use
	packed, err := proto.PackRoot(rep)
	if err != nil {
		return serrors.WrapStr("Error packing reply to insert into db", err)
	}
	_, err = sqlite.Db.InsertNewMSToken(context.Background(), packed)
	if err != nil {
		return serrors.WrapStr("Error storing MS token into db", err)
	}

	//Add the pushed prefix into the database
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

func registerSigner(msgType infra.MessageType) error {
	sigcrypt := &sigcrypto.SIGSigner{}
	err := sigcrypt.Init(context.Background(), sigcmn.Msgr, sigcmn.IA, sigcrypto.CfgDir)
	if err != nil {
		return serrors.WrapStr("Unable to initialize sig crypto", err)
	}
	signer, err := sigcrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	sigcmn.Msgr.UpdateSigner(signer, []infra.MessageType{msgType})
	return nil
}
