package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type SignedASEntry struct {
	Blob common.RawBytes
	Sign *proto.SignS
}

func NewSignedASEntry(blob common.RawBytes, sign *proto.SignS) *SignedASEntry {
	return &SignedASEntry{Blob: blob, Sign: sign}
}

func (p *SignedASEntry) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *SignedASEntry) String() string {
	return fmt.Sprintf("SignedASEntry: %s %s", p.Blob, p.Sign)
}

type SignedMSList struct {
	Timestamp uint64
	ASEntries []SignedASEntry `capnp:"asEntries"`
	MSIA      string          `capnp:"msIA"`
}

func NewSignedMSList(timestamp uint64, asEntry []SignedASEntry,
	msIA string) *SignedMSList {

	return &SignedMSList{Timestamp: timestamp, ASEntries: asEntry, MSIA: msIA}
}

func (p *SignedMSList) ProtoId() proto.ProtoIdType {
	return proto.SignedMSList_TypeID
}

func (p *SignedMSList) String() string {
	s := ""
	for _, asEntry := range p.ASEntries {
		s = s + asEntry.String()
	}
	return fmt.Sprintf("%d %s %s", p.Timestamp, s, p.MSIA)
}
