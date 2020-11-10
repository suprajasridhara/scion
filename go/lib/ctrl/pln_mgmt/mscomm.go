package pln_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type PlnListReq struct {
	Action string
}

func NewPlnListReq(action string) *PlnListReq {
	return &PlnListReq{Action: action}
}

func (p *PlnListReq) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnListReq) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnListReq) String() string {
	return fmt.Sprintf("%s ", p.Action)
}

type PlnListEntry struct {
	PCNId string `capnp:"pcnId"`
	IA    uint64 `capnp:"ia"`
	Raw   []byte `capnp:"raw"`
}

func NewPlnListEntry(pcnId string, ia uint64, raw []byte) *PlnListEntry {
	return &PlnListEntry{PCNId: pcnId, IA: ia, Raw: raw}
}

func (p *PlnListEntry) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnListEntry) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnListEntry) String() string {
	return fmt.Sprintf("%s %d %s", p.PCNId, p.IA, p.Raw)
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
