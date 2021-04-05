package ms_mgmt

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

// union represents the contents of the unnamed capnp union.
type union struct {
	Which        proto.MS_Which
	AsActionReq  *ASMapEntry
	AsActionRep  *MSRepToken
	FullMapReq   *FullMapReq
	FullMapRep   *FullMapRep
	SignedMSList *SignedMSList
}

func (u *union) set(c proto.Cerealizable) error {
	switch p := c.(type) {
	case *ASMapEntry:
		u.Which = proto.MS_Which_asActionReq
		u.AsActionReq = p
	case *MSRepToken:
		u.Which = proto.MS_Which_asActionRep
		u.AsActionRep = p
	case *FullMapReq:
		u.Which = proto.MS_Which_fullMapReq
		u.FullMapReq = p
	case *FullMapRep:
		u.Which = proto.MS_Which_fullMapRep
		u.FullMapRep = p
	case *SignedMSList:
		u.Which = proto.MS_Which_signedMSList
		u.SignedMSList = p
	default:
		return common.NewBasicError("Unsupported MS ctrl union type (set)", nil,
			"type", common.TypeOf(c))
	}
	return nil
}

func (u *union) get() (proto.Cerealizable, error) {
	switch u.Which {
	case proto.MS_Which_asActionReq:
		return u.AsActionReq, nil
	case proto.MS_Which_asActionRep:
		return u.AsActionRep, nil
	case proto.MS_Which_fullMapReq:
		return u.FullMapReq, nil
	case proto.MS_Which_fullMapRep:
		return u.FullMapRep, nil
	case proto.MS_Which_signedMSList:
		return u.SignedMSList, nil
	}
	return nil, common.NewBasicError("Unsupported MS ctrl union type (get)", nil,
		"type", u.Which)
}
