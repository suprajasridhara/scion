package pcn_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type MSListRep struct {
	SignedMSList []byte `capnp:"signedMSList"`
	CommitId     string
	Timestamp    uint64
}

func NewMSListRep(signedMSList []byte, commitId string, timestamp uint64) *MSListRep {
	return &MSListRep{SignedMSList: signedMSList, CommitId: commitId, Timestamp: timestamp}
}

func (p *MSListRep) ProtoId() proto.ProtoIdType {
	return proto.PCN_TypeID
}

func (p *MSListRep) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *MSListRep) String() string {
	return fmt.Sprintf("%s %s %d", string(p.SignedMSList), p.CommitId, p.Timestamp)
}

type NodeListEntryRequest struct {
	Query string
}

func NewNodeListEntryRequest(query string) *NodeListEntryRequest {
	return &NodeListEntryRequest{Query: query}
}

func (p *NodeListEntryRequest) ProtoId() proto.ProtoIdType {
	return proto.PCN_TypeID
}

func (p *NodeListEntryRequest) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *NodeListEntryRequest) String() string {
	return fmt.Sprintf("%s ", p.Query)
}
