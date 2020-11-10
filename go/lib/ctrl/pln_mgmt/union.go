package pln_mgmt

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

// union represents the contents of the unnamed capnp union.
type union struct {
	Which     proto.PLN_Which
	MsListReq *PlnListReq
	PlnList   *PlnList `capnp:"plnList"`
}

func (u *union) set(c proto.Cerealizable) error {
	switch p := c.(type) {
	case *PlnListReq:
		u.Which = proto.PLN_Which_msListReq
		u.MsListReq = p
	case *PlnList:
		u.Which = proto.PLN_Which_plnList
		u.PlnList = p
	default:
		return common.NewBasicError("Unsupported PLN ctrl union type (set)", nil,
			"type", common.TypeOf(c))
	}
	return nil
}

func (u *union) get() (proto.Cerealizable, error) {
	switch u.Which {
	case proto.PLN_Which_plnList:
		return u.PlnList, nil
	case proto.PLN_Which_msListReq:
		return u.MsListReq, nil
	}
	return nil, common.NewBasicError("Unsupported PLN ctrl union type (get)", nil,
		"type", u.Which)
}
