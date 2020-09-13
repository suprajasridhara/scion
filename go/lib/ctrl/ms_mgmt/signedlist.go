package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/proto"
)

type SignedMSList struct {
	Timestamp uint64
	PCNId     string `capnp:"pcnId"`
	AsEntry   ctrl.SignedPld
}

func NewSignedMSList(timestamp uint64, pcnId string, asEntry ctrl.SignedPld) *SignedMSList {
	return &SignedMSList{Timestamp: timestamp, PCNId: pcnId, AsEntry: asEntry}
}

func (p *SignedMSList) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *SignedMSList) String() string {
	return fmt.Sprintf("%d %s %s", p.Timestamp, p.PCNId, p.AsEntry.String())
}
