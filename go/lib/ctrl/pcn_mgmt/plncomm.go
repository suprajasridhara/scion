package pcn_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/pln_mgmt"
	"github.com/scionproto/scion/go/proto"
)

type AddPLNEntryRequest struct {
	Entry pln_mgmt.PlnListEntry
}

func NewAddPLNEntryRequest(entry pln_mgmt.PlnListEntry) *AddPLNEntryRequest {
	return &AddPLNEntryRequest{Entry: entry}
}

func (p *AddPLNEntryRequest) ProtoId() proto.ProtoIdType {
	return proto.PCN_TypeID
}

func (p *AddPLNEntryRequest) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *AddPLNEntryRequest) String() string {
	return fmt.Sprintf("%s ", p.Entry.String)
}
