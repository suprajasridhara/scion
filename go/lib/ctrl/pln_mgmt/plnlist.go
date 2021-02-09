package pln_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type PlnListEntry struct {
	PGNId string `capnp:"pgnId"`
	IA    uint64 `capnp:"ia"`
	Raw   []byte `capnp:"raw"`
}

func NewPlnListEntry(pgnId string, ia uint64, raw []byte) *PlnListEntry {
	return &PlnListEntry{PGNId: pgnId, IA: ia, Raw: raw}
}

func (p *PlnListEntry) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnListEntry) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnListEntry) String() string {
	return fmt.Sprintf("%s %d %s", p.PGNId, p.IA, p.Raw)
}

type PlnList struct {
	L []PlnListEntry `capnp:"l"`
}

func NewPlnList(l []PlnListEntry) *PlnList {
	return &PlnList{L: l}
}

func (p *PlnList) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnList) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnList) String() string {
	var s []string
	for _, l := range p.L {
		s = append(s, l.String())
	}
	return fmt.Sprintf("%v", s)
}
