package plncomm

import (
	"context"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/pkg/trust/compat"
)

//PLNAddr address of the PLN that the MS is configured to connect to
var PLNAddr addr.IA

//PCN pcn objects that are returned from the PLN
type PCN struct {
	//PCNId is the id of the PCN In the IA it is deployed
	PCNId string
	//PCNIA ia of the PCN
	PCNIA addr.IA
}

/*GetPlnList The Mapping Service sends the request using the infra.Messenger instance in msmsgr package and
verifies the origin of the response before processing it. It then returns the processed list of
PCN Id and IA objects to the calling function
*/
func GetPlnList(ctx context.Context) ([]PCN, error) {
	address := &snet.SVCAddr{IA: PLNAddr, SVC: addr.SvcPLN}

	plnListReq := pln_mgmt.NewPlnListReq("request")
	pld, err := pln_mgmt.NewPld(1, plnListReq)
	if err != nil {
		return nil, serrors.WrapStr("Error creating pln_mgmt pld", err)
	}

	signedPld, err := msmsgr.Msgr.GetPlnList(ctx, pld, address, 1)
	if err != nil {
		return nil, serrors.WrapStr("Error getting plnList from messenger", err)
	}
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: PLNAddr, Engine: e}

	verifiedPayload, err := signedPld.GetVerifiedPld(context.Background(),
		compat.Verifier{Verifier: verifier})

	if err != nil {
		return nil, serrors.WrapStr("Error getting verified payload", err)
	}
	plnList := verifiedPayload.Pln.PlnList

	pcns := []PCN{}

	for _, plnListEntry := range plnList.L {
		pcn := PCN{PCNId: plnListEntry.PCNId, PCNIA: addr.IAInt(plnListEntry.IA).IA()}
		pcns = append(pcns, pcn)
	}
	//Signature from PLN is validated, the list is now authenticated.

	return pcns, nil
}
