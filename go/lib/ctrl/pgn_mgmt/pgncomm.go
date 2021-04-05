package pgn_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type PGNList struct {
	//L should be a list of SignedPlds in bytes form.
	L         []common.RawBytes
	Timestamp uint64
}

func NewPGNList(l []common.RawBytes, timestamp uint64) *PGNList {
	return &PGNList{L: l, Timestamp: timestamp}
}

func (p *PGNList) ProtoId() proto.ProtoIdType {
	return proto.PGN_TypeID
}

func (p *PGNList) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PGNList) String() string {
	str := ""
	return fmt.Sprintf("%s %d", str, p.Timestamp)
}
