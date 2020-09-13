package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type SignedAsEntry struct {
	Blob common.RawBytes
	Sign *proto.SignS
}

func NewSignedAsEntry(blob common.RawBytes, sign *proto.SignS) *SignedAsEntry {
	return &SignedAsEntry{Blob: blob, Sign: sign}
}

func (p *SignedAsEntry) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *SignedAsEntry) String() string {
	return fmt.Sprintf("SignedAsEntry: %s %s", p.Blob, p.Sign)
}

type SignedMSList struct {
	Timestamp uint64
	PCNId     string          `capnp:"pcnId"`
	AsEntries []SignedAsEntry `capnp:"asEntries"`
}

func NewSignedMSList(timestamp uint64, pcnId string, asEntry []SignedAsEntry) *SignedMSList {
	return &SignedMSList{Timestamp: timestamp, PCNId: pcnId, AsEntries: asEntry}
}

func (p *SignedMSList) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *SignedMSList) String() string {
	s := ""
	for _, asEntry := range p.AsEntries {
		s = s + asEntry.String()
	}
	return fmt.Sprintf("%d %s %s", p.Timestamp, p.PCNId, s)
}
