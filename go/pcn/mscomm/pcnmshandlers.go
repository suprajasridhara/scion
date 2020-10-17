package mscomm

import (
	"context"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/internal/sqlite"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

//TODO (supraja): move this and make it configurable
const ms_list_valid_time = 1000000 * time.Minute

type MSListHandler struct {
}

type NodeListEntryReqHandler struct {
}

func (m MSListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: MSListHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)

	//Verify AS signature
	message := r.FullMessage.(*ctrl.SignedPld)
	e := pcncrypto.PCNEngine{Msgr: pcnmsgr.Msgr, IA: pcnmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	// msmsgr.Msgr.UpdateVerifier(verifier)
	err := verifier.Verify(ctx, message.Blob, message.Sign)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
		log.Error("Certificate verification failed!", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	//Verify timestamp. List should be valid at the time this check is performed
	signedMSList := r.Message.(*ms_mgmt.SignedMSList)
	_, err = isValidMSList(r.Peer.(*snet.UDPAddr).IA, *signedMSList)
	if err != nil {
		log.Error("Error validating list", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	//Persist (for now persist full entry, TODO (supraja): updates and commit ids)
	fm := r.FullMessage.(*ctrl.SignedPld)
	packed, err := proto.PackRoot(fm)
	commitId := generateCommitID()
	err = persistMSList(context.Background(), packed, commitId, signedMSList.MSIA, signedMSList.Timestamp)
	if err != nil {
		log.Error("Error persisting list", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	//send back signed response
	pcncrypt := &pcncrypto.PCNSigner{}
	pcncrypt.Init(ctx, pcnmsgr.Msgr, pcnmsgr.IA, pcncrypto.CfgDir)
	signer, err := pcncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		log.Error("error getting signer")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}
	pcnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PushMSListReply})

	//TODO (supraja): change timestamp to be the max amount of time pcn needs to broadcast the list.
	msListRep := pcn_mgmt.NewMSListRep(packed, commitId, uint64(time.Now().Unix()))
	pld, err := pcn_mgmt.NewPld(1, msListRep)
	if err != nil {
		log.Error("Error getting pcn_mgmt.pld")
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pcnmsgr.Msgr.SignedMSListRep(context.Background(), pld, r.Peer, r.ID)
	return nil
}

func isValidMSList(peerIA addr.IA, l ms_mgmt.SignedMSList) (bool, error) {
	//Validate timestamp
	timestamp := time.Unix(int64(l.Timestamp), 0)

	if !timestamp.Add(ms_list_valid_time).After(time.Now()) {
		return false, serrors.New("Invalid timstamp in SignedMSList", "")
	}

	//Validate that the entries are in the same ISD. For this validate the signature on the ASEntry first
	for _, asEntry := range l.AsEntries {
		// asME := &ms_mgmt.ASMapEntry{}
		// err := proto.ParseFromRaw(asME, asEntry.Blob)
		spld := &ctrl.Pld{}
		err := proto.ParseFromRaw(spld, asEntry.Blob)
		if err != nil {
			log.Error("Error decerializing", err)
			return false, err
		}
		print(spld.Ms.String())
		mapEntry := spld.Ms.AsActionReq
		mapEntryIA, err := addr.IAFromString(mapEntry.Ia)
		if err != nil {
			log.Error("Error getting Ia", err)
			return false, err
		}
		if mapEntryIA.I != peerIA.I {
			return false, serrors.New("Invalid mapEntry: ia "+mapEntryIA.String(), "")
		}
		// print(asME.Ia)
	}

	if l.PCNId != pcnmsgr.Id {
		return false, serrors.New("Invalid PCN Id", "")
	}

	//validate the AS entries are the AS for which the MS is authoritative for
	return true, nil
}

func (n NodeListEntryReqHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: NodeListEntryReqHandler.Handle")
	return nil
}

func persistMSList(ctx context.Context, signedMSList []byte, commitId string, msIA string, timestamp uint64) error {
	_, err := sqlite.Db.InsertNewNodeListEntry(ctx, signedMSList, commitId, msIA, timestamp)
	return err
}
func generateCommitID() string {
	//TODO (supraja): implement this correctly when doing the update messages
	return "1234"
}
