package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type OkMessage struct {
	Ok string
}

func NewOkMessage(ok string) *OkMessage {
	return &OkMessage{Ok: ok}
}

func (p *OkMessage) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *OkMessage) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *OkMessage) String() string {
	return fmt.Sprintf("%s", p.Ok)
}
