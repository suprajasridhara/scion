package pcncomm

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pcn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite3"
	"github.com/scionproto/scion/go/ms/plncomm"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

//TODO (supraja): read from config file
//valid time in hours
const ms_list_valid_time = 100000

func SendSignedList(ctx context.Context, interval time.Duration) {
	pushSignedPrefix(ctx)
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			pushSignedPrefix(ctx)
		}
	}
}

func pushSignedPrefix(ctx context.Context) {
	logger := log.FromCtx(ctx)

	//signed ASMapEntries in the form of SignedPld
	asEntries, err := sqlite3.Db.GetNewEntries(context.Background())
	if err != nil {
		logger.Error("could not get entries from DB", "Err: ", err)
	}

	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		logger.Error("error getting signer", "Err: ", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.PushMSListRequest})

	entries := []ms_mgmt.SignedAsEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedAsEntry(asEntry.Blob, asEntry.Sign)
		entries = append(entries, *entry)
	}

	timestamp := time.Now()

	// pcnIA := pcns[0].PCNIA
	pcn := getRandomPCN(ctx)
	address := &snet.SVCAddr{IA: pcn.PCNIA, SVC: addr.SvcPCN}
	req := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), pcn.PCNId, entries, msmsgr.IA.String())
	print("Timestamp in MSLIST : ", timestamp.Unix())
	pld, err := ms_mgmt.NewPld(1, req)
	if err != nil {
		logger.Error("error forming ms_mgmt payload", "Err: ", err)
	}

	//TODO_Q (supraja): generate random id?
	reply, err := msmsgr.Msgr.SendSignedMSList(ctx, pld, address, 123)

	if err != nil {
		logger.Error("error getting reply from PCN", "Err: ", err)
	}

	//Validate PCN signature
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: address.IA, Engine: e}
	err = verifier.Verify(ctx, reply.Blob, reply.Sign)

	if err != nil {
		logger.Error("error verifying sign for PCN rep", "Err: ", err)
	}

	packed, err := proto.PackRoot(reply)
	if err != nil {
		logger.Error("Error packing reply", "Err: ", err)
	}
	_, err = sqlite3.Db.InsertPCNRep(context.Background(), packed)

	if err != nil {
		logger.Error("error persisting PCN rep", "Err: ", err)
	}

}

func PullFullNodeList(ctx context.Context, interval time.Duration) {
	pushSignedPrefix(ctx)
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			PullNodeListEntry(ctx, "") //"" is considered wildcard.
		}
	}
}

func PullNodeListEntry(ctx context.Context, query string) {
	logger := log.FromCtx(ctx)
	mscrypt := &mscrypto.MSSigner{}
	err := mscrypt.Init(ctx, msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	if err != nil {
		logger.Error("error mscrypt init", "Err: ", err)
	}
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		logger.Error("error getting signer", "Err: ", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.NodeListEntryRequest})
	req := pcn_mgmt.NewNodeListEntryRequest(query)
	pld, err := pcn_mgmt.NewPld(1, req)
	if err != nil {
		logger.Error("error constructing payload", "Err: ", err)
	}
	pcn := getRandomPCN(ctx)
	address := &snet.SVCAddr{IA: pcn.PCNIA, SVC: addr.SvcPCN}

	spld, err := msmsgr.Msgr.SendNodeListRequest(ctx, pld, address, rand.Uint64())
	if err != nil {
		logger.Error("Eror sending node list request to PCN ", "Err: ", err)
	}
	//verify AS signatures
	err = verifyASSignature(context.Background(), spld, pcn.PCNIA)
	if err != nil {
		logger.Error("error verifying AS signature on response from PCN", "Err: ", err)
	}

	//parse signed payload to get the nodelist
	cpld := &ctrl.Pld{}
	err = proto.ParseFromRaw(cpld, spld.Blob)
	if err != nil {
		logger.Error("error parsing payload", "Err: ", err)
	}

	nodeList := cpld.Pcn.NodeList
	validateAndPersistNLEs(nodeList.L)
	print(nodeList.L[0].CommitId)

}

func getRandomPCN(ctx context.Context) plncomm.PCN {
	logger := log.FromCtx(ctx)
	pcns, err := plncomm.GetPlnList(context.Background())
	if err != nil {
		logger.Error("error getting pcns", "Err: ", err)
	}
	//pick a random pcn to send signed list to
	randomIndex := rand.Intn(len(pcns))
	return pcns[randomIndex]
}

func verifyASSignature(ctx context.Context, message *ctrl.SignedPld, IA addr.IA) error {
	//Verify AS signature
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: IA, Engine: e}
	return verifier.Verify(ctx, message.Blob, message.Sign)
}

func validateAndPersistNLEs(nodeListEntries []pcn_mgmt.NodeListEntry) {
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
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
		if uint64(time.Now().Unix())-msPld.Ms.PushMSListReq.Timestamp >
			uint64(ms_list_valid_time*time.Hour) {
			log.Error("msList entry too old. Reject", err)
			continue
		}

		//process and persist here
		asEntries := msPld.Ms.PushMSListReq.AsEntries

		for _, asEntry := range asEntries {
			cpld := &ctrl.Pld{}
			err = proto.ParseFromRaw(cpld, asEntry.Blob)
			if err != nil {
				log.Error("Error decerializing asEntries", err)
				continue
			}
			asMapEntry := cpld.Ms.AsActionReq

			//TODO (supraja): validate RPKI signature here for every prefix in the ASEntry?

			//verify that the IA In the map originated the entry by checking the signature
			ia, err := addr.IAFromString(asMapEntry.Ia)
			verifier = trust.Verifier{BoundIA: ia, Engine: e}
			//err = verifier.Verify(context.Background(), asEntry.Blob, asEntry.Sign)
			if err != nil {
				log.Error("Certificate verification failed for ASMapEntry source AS", err)
				continue
			}

			for _, prefix := range asMapEntry.Ip {
				//query by IA, IP

				oldMapEntries, err := sqlite3.Db.GetFullMapEntryByIp(context.Background(), prefix)
				if err != nil {
					log.Error("Error getting old entries for IP from DB", err)
					continue
				}

				for _, oldMapEntry := range oldMapEntries {
					/*if the rpki signature was validated for the new entry then it is the
					entry with the most trust, delete all other entries*/
					_, err := sqlite3.Db.DeleteFullMapEntryById(context.Background(),
						oldMapEntry.Id)
					if err != nil {
						log.Error("Error deleting old entry")
					}
				}

				_, err = sqlite3.Db.InsertFullMapEntry(context.Background(),
					sqlite3.FullMapRow{IP: sql.NullString{String: prefix, Valid: true},
						IA:        sql.NullString{String: asMapEntry.Ia, Valid: true},
						Timestamp: int(asMapEntry.Timestamp)})

				if err != nil {
					log.Error("Error inserting entry into db", err)
					continue
				}
			}

		}

	}
}
