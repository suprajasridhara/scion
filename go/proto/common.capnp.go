// Code generated by capnpc-go. DO NOT EDIT.

package proto

import (
	capnp "zombiezen.com/go/capnproto2"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type LinkType uint16

// LinkType_TypeID is the unique identifier for the type LinkType.
const LinkType_TypeID = 0x916c98f48c9bbb64

// Values of LinkType.
const (
	LinkType_unset  LinkType = 0
	LinkType_core   LinkType = 1
	LinkType_parent LinkType = 2
	LinkType_child  LinkType = 3
	LinkType_peer   LinkType = 4
)

// String returns the enum's constant name.
func (c LinkType) String() string {
	switch c {
	case LinkType_unset:
		return "unset"
	case LinkType_core:
		return "core"
	case LinkType_parent:
		return "parent"
	case LinkType_child:
		return "child"
	case LinkType_peer:
		return "peer"

	default:
		return ""
	}
}

// LinkTypeFromString returns the enum value with a name,
// or the zero value if there's no such value.
func LinkTypeFromString(c string) LinkType {
	switch c {
	case "unset":
		return LinkType_unset
	case "core":
		return LinkType_core
	case "parent":
		return LinkType_parent
	case "child":
		return LinkType_child
	case "peer":
		return LinkType_peer

	default:
		return 0
	}
}

type LinkType_List struct{ capnp.List }

func NewLinkType_List(s *capnp.Segment, sz int32) (LinkType_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return LinkType_List{l.List}, err
}

func (l LinkType_List) At(i int) LinkType {
	ul := capnp.UInt16List{List: l.List}
	return LinkType(ul.At(i))
}

func (l LinkType_List) Set(i int, v LinkType) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type ServiceType uint16

// ServiceType_TypeID is the unique identifier for the type ServiceType.
const ServiceType_TypeID = 0xe2522d291bd06774

// Values of ServiceType.
const (
	ServiceType_unset ServiceType = 0
	ServiceType_bs    ServiceType = 1
	ServiceType_ps    ServiceType = 2
	ServiceType_cs    ServiceType = 3
	ServiceType_sb    ServiceType = 4
	ServiceType_ds    ServiceType = 5
	ServiceType_br    ServiceType = 6
	ServiceType_sig   ServiceType = 7
	ServiceType_hps   ServiceType = 8
	ServiceType_ms    ServiceType = 9
	ServiceType_pln   ServiceType = 10
	ServiceType_pgn   ServiceType = 11
)

// String returns the enum's constant name.
func (c ServiceType) String() string {
	switch c {
	case ServiceType_unset:
		return "unset"
	case ServiceType_bs:
		return "bs"
	case ServiceType_ps:
		return "ps"
	case ServiceType_cs:
		return "cs"
	case ServiceType_sb:
		return "sb"
	case ServiceType_ds:
		return "ds"
	case ServiceType_br:
		return "br"
	case ServiceType_sig:
		return "sig"
	case ServiceType_hps:
		return "hps"
	case ServiceType_ms:
		return "ms"
	case ServiceType_pln:
		return "pln"
	case ServiceType_pgn:
		return "pgn"
	default:
		return ""
	}
}

// ServiceTypeFromString returns the enum value with a name,
// or the zero value if there's no such value.
func ServiceTypeFromString(c string) ServiceType {
	switch c {
	case "unset":
		return ServiceType_unset
	case "bs":
		return ServiceType_bs
	case "ps":
		return ServiceType_ps
	case "cs":
		return ServiceType_cs
	case "sb":
		return ServiceType_sb
	case "ds":
		return ServiceType_ds
	case "br":
		return ServiceType_br
	case "sig":
		return ServiceType_sig
	case "hps":
		return ServiceType_hps
	case "ms":
		return ServiceType_ms
	case "pln":
		return ServiceType_pln
	case "pgn":
		return ServiceType_pgn
	default:
		return 0
	}
}

type ServiceType_List struct{ capnp.List }

func NewServiceType_List(s *capnp.Segment, sz int32) (ServiceType_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return ServiceType_List{l.List}, err
}

func (l ServiceType_List) At(i int) ServiceType {
	ul := capnp.UInt16List{List: l.List}
	return ServiceType(ul.At(i))
}

func (l ServiceType_List) Set(i int, v ServiceType) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

const schema_fa01960eced2b529 = "x\xda\x12\xc8t`\x12d\xdd\xce\xc0\x10\xc8\xc1\xca\xf6" +
	"?e\xf7\xec\x9e/3r&2\x08\xf22\xfd\xd7\xdc" +
	"z\xe9\x1c\xdf4\xc6_\x0c\x0c\x8c\x82\x8e\x9b\x04=\xd9" +
	"\x19\x18\x04]\xeb\x19\x18\xff\x97\xa4_\x90\xd6\xd4\x0dz" +
	"\x84\xa1\xaa\xf2\x94`+;\x03\x83a\xa3:#\x03\xe3" +
	"\xff\xe4\xfc\xdc\xdc\xfc<\xbdd\xc6\xc4\x82\xbc\x02+\x9f" +
	"\xcc<\xf9\xec\x90\xca\x82\xd4\x00F\xc6@\x11F&\x06" +
	"\x06AS#\x06\x06FFA]-\x06\x06F&A" +
	"U+\x06\x06FfAY\x90 \x8b\xa0\xa8\x16\x03\x83" +
	"|i^qj\x09\x7fr~Q\xaa}AbQj" +
	"^\x89|rFfN\x0a\x7fAjj\x11\xdcx&" +
	"\xb0\xf1\xc1\xa9Ee\x99\xc9\xa9 \x0b\x18\x18@V\x18" +
	"\x80\xad\xe8\x84X\xd1(\x05\xb6\xa2R\x0alE\xa1\x14" +
	"\xd8\x8aL\x10\xc5*\x98\x08\xa2\xd8\x04#A\x14\xbb`" +
	"\xa0\x12\x03\x03#\x87\xa0'\x88\xe2\x14t\x04\x09r\x09" +
	"Z\x82x\xdc\x82\x86J071'\x153\x17\x143" +
	"'\x173\x17'1\xa7\x143'\x15\xb1\x17g\xa6\xb3" +
	"g\x14\x143\xe7\x16\xb3\x17\xe4\xe4\xb1\x17\xa4\xe7\x01\x02" +
	"\x00\x00\xff\xff\x95PR\x0e"
const schema_fa01960eced2b529 = "x\xda\x12Ht`\x12d\xdd\xce\xc0\x10\xc8\xc1\xca\xf6" +
	"?e\xf7\xec\x9e/3r&2\x08\xf22\xfd\xd7\xdc" +
	"z\xe9\x1c\xdf4\xc6_\x0c\x0c\x8c\x82\x8e\x9b\x04=\xd9" +
	"\x19\x18\x04]\xeb\x19\x18\xff\x97\xa4_\x90\xd6\xd4\x0dz" +
	"\x84\xa1\xaa\xf2\x94`+HU\xe3w\x06\xc6\xff\xc9\xf9" +
	"\xb9\xb9\xf9yz\xc9\x8c\x89\x05y\x05V>\x99y\xf2" +
	"\xd9!\x95\x05\xa9\x01\x8c\x8c\x81\"\x8cL\x0c\x0c\x82\xa6" +
	"F\x0c\x0c\x8c\x8c\x82\xbaZ\x0c\x0c\x8cL\x82\xaaV\x0c" +
	"\x0c\x8c\xcc\x82\xb2 A\x16AQ-\x06\x06\xf9\xd2\xbc" +
	"\xe2\xd4\x12\xfe\xe4\xfc\xa2T\xfb\x82\xc4\xa2\xd4\xbc\x12\xf9" +
	"\xe4\x8c\xcc\x9c\x14\xfe\x82\xd4\xd4\"\xb8\xf1L`\xe3\x83" +
	"S\x8b\xca2\x93SA\x1600\x80\xac\xd0\x00[Q" +
	"\x08\xb1\"S\x0alE\xa2\x14\xd8\x8aH)\xb0\x15\x81" +
	" \x8aU\xd0\x13D\xb1\x09:\x82(vAK%\x06" +
	"\x06F\x0eAC\x10\xc5)\xa8)\x05s\x05sR1" +
	"sA1sr1sq\x12sJ1sR\x11{" +
	"qf:{FA1sn1 \x00\x00\xff\xffm" +
	"\x13N9"

func init() {
	schemas.Register(schema_fa01960eced2b529,
		0x916c98f48c9bbb64,
		0xe2522d291bd06774)
}
