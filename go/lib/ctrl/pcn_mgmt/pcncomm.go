package pcn_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type NodeList struct {
	L         []NodeListEntry
	Timestamp uint64
}

func NewNodeList(l []NodeListEntry, timestamp uint64) *NodeList {
	return &NodeList{L: l, Timestamp: timestamp}
}

func (n *NodeList) ProtoId() proto.ProtoIdType {
	return proto.PCN_TypeID
}

func (n *NodeList) String() string {
	str := ""
	for _, l := range n.L {
		str = str + l.String()
	}
	return fmt.Sprintf("%s %d", str, n.Timestamp)
}

type NodeListEntry struct {
	SignedMSList common.RawBytes
	CommitId     string
}

func NewNodeListEntry(signedMSList common.RawBytes, commitId string) *NodeListEntry {
	return &NodeListEntry{SignedMSList: signedMSList, CommitId: commitId}
}

func (n *NodeListEntry) ProtoId() proto.ProtoIdType {
	return proto.PCN_TypeID
}

func (n *NodeListEntry) String() string {
	return fmt.Sprintf("%s %s", n.SignedMSList, n.CommitId)
}
