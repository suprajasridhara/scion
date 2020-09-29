package pcn_mgmt

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

// union represents the contents of the unnamed capnp union.
type union struct {
	Which              proto.PCN_Which
	AddPLNEntryRequest *AddPLNEntryRequest
	MSListRep          *MSListRep `capnp:"msListRep"`
}

func (u *union) set(c proto.Cerealizable) error {
	switch p := c.(type) {
	case *AddPLNEntryRequest:
		u.Which = proto.PCN_Which_addPLNEntryRequest
		u.AddPLNEntryRequest = p
	case *MSListRep:
		u.Which = proto.PCN_Which_msListRep
	default:
		return common.NewBasicError("Unsupported PCN ctrl union type (set)", nil,
			"type", common.TypeOf(c))
	}
	return nil
}

func (u *union) get() (proto.Cerealizable, error) {
	switch u.Which {
	case proto.PCN_Which_addPLNEntryRequest:
		return u.AddPLNEntryRequest, nil

	case proto.PCN_Which_msListRep:
		return u.MSListRep, nil

	}
	return nil, common.NewBasicError("Unsupported PLN ctrl union type (get)", nil,
		"type", u.Which)
}