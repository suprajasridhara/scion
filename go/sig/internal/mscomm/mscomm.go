package mscomm

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
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
	ia := addr.IA{
		I: sigcmn.IA.I,
		A: sigcmn.CoreASes[0],
	}
	addr := &snet.SVCAddr{IA: ia, SVC: addr.SvcMS}
	timestamp := uint64(time.Now().UnixNano())
	asEntry := ms_mgmt.NewASMapEntry([]string{ip}, sigcmn.IA.String(), timestamp, ADD_AS_ENTRY)

	sigcrypt := &sigcrypto.SIGSigner{}
	sigcrypt.Init(context.Background(), sigcmn.Msgr, sigcmn.IA, sigcrypto.CfgDir)
	signer, err := sigcrypt.SignerGen.Generate(context.Background())
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

	//Verify MS Signature
	e := sigcrypto.SIGEngine{Msgr: sigcmn.Msgr, IA: sigcmn.IA}
	verifier := trust.Verifier{BoundIA: ia, Engine: e}
	err = verifier.Verify(context.Background(), rep.Blob, rep.Sign)
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
