package pcncomm

import (
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/internal/sqlite"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
	"golang.org/x/net/context"
)

type NodeListHandler struct {
}

//TODO (supraja): read from config file
const ms_list_valid_time = 5

func (n NodeListHandler) Handle(r *infra.Request) *infra.HandlerResult {
	log.Info("Entering: NodeListHandler.Handle")
	ctx := r.Context()
	requester := r.Peer.(*snet.UDPAddr)

	//Verify node list signature by pcn on the list
	message := r.FullMessage.(*ctrl.SignedPld)
	e := pcncrypto.PCNEngine{Msgr: pcnmsgr.Msgr, IA: pcnmsgr.IA}
	verifier := trust.Verifier{BoundIA: requester.IA, Engine: e}
	err := verifier.Verify(ctx, message.Blob, message.Sign)
	rw, _ := infra.ResponseWriterFromContext(ctx)
	sendAck := messenger.SendAckHelper(ctx, rw)
	if err != nil {
		log.Error("Certificate verification failed!", err)
		sendAck(proto.Ack_ErrCode_reject, err.Error())
		return nil
	}

	pld := &ctrl.Pld{}
	err = proto.ParseFromRaw(pld, message.Blob)
	if err != nil {
		log.Error("Error decerializing control payload", err)
		return nil
	}
	//for each entry verify AS signature and timestamp
	validateAndPersistNodeListEntries(pld.Pcn.NodeList.L, e)

	return nil
}

func validateAndPersistNodeListEntries(nodeListEntries []pcn_mgmt.NodeListEntry, e pcncrypto.PCNEngine) {
	for _, nodeListEntry := range nodeListEntries {
		//verify ms signature on msList
		spld := &ctrl.SignedPld{}
		err := proto.ParseFromRaw(spld, nodeListEntry.SignedMSList)
		if err != nil {
			log.Error("Error decerializing nodeListEntry", err)
			continue
		}

		msPld := &ctrl.Pld{}
		err = proto.ParseFromRaw(msPld, spld.Blob)
		if err != nil {
			log.Error("Error decerializing msPld", err)
			continue
		}
		msIA, err := addr.IAFromString(msPld.Ms.PushMSListReq.MSIA)
		if err != nil {
			log.Error("Error getting ms IA from string", err)
			continue
		}

		verifier := trust.Verifier{BoundIA: msIA, Engine: e}
		err = verifier.Verify(context.Background(), spld.Blob, spld.Sign)
		if err != nil {
			log.Error("Certificate verification failed for MS!", err)
			continue
		}

		//verify timestamp
		// timeNow := time.Now().Unix()
		// print(timeNow)
		//compTimestamp := time.Now().Add(ms_list_valid_time * time.Minute).UTC().Unix()
		if uint64(time.Now().Unix())-msPld.Ms.PushMSListReq.Timestamp > 300 {
			log.Error("msList entry too old. Reject", err)
			continue
		}

		//persist NodeListEntries

		fullNodeList, err := sqlite.Db.GetFullNodeList(context.Background())
		if err != nil {
			log.Error("error reading list from db", err)
		}
		insert := true
		update := false
		//check if entry exists with the same commit ID, replace only if not
		for _, nle := range fullNodeList {
			if nle.MSIA.String == msIA.String() && nle.CommitId.String == nodeListEntry.CommitId { //if msIA and commitId both match it should be the same list. No need to insert
				insert = false
				break
			} else if nle.MSIA.String == msIA.String() && uint64(nle.Timestamp) > msPld.Ms.PushMSListReq.Timestamp { //if msIA is the same and the timestamp in the db is newer then there is no need to update it
				insert = false
				break
			} else if nle.MSIA.String == msIA.String() {
				update = true
				break
			}
		}
		if update {
			log.Info("Updating MSList in DB. MSIA: " + msIA.String())
			sqlite.Db.UpdateNodeListEntry(context.Background(), nodeListEntry.SignedMSList, nodeListEntry.CommitId, msIA.String(), msPld.Ms.PushMSListReq.Timestamp)
		} else if insert {
			log.Info("Inserting MSList to DB. MSIA: "+msIA.String(), nil)
			sqlite.Db.InsertNewNodeListEntry(context.Background(), nodeListEntry.SignedMSList, nodeListEntry.CommitId, msIA.String(), msPld.Ms.PushMSListReq.Timestamp)
		}
	}
}
