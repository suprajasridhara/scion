package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type ASMapEntry struct {
	Ia        string
	Ip        []string
	Timestamp uint64
	Action    string
}

func NewASMapEntry(ip []string, ia string, timestamp uint64, action string) *ASMapEntry {
	return &ASMapEntry{Ip: ip, Ia: ia, Timestamp: timestamp, Action: action}
}

func (p *ASMapEntry) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *ASMapEntry) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *ASMapEntry) String() string {
	return fmt.Sprintf("%s %s %d %s", p.Ip, p.Ia, p.Timestamp, p.Action)
}

type MSRepToken struct {
	SignedASEntry []byte
	Timestamp     uint64
}

func NewMSRepToken(signedASEntry []byte, timestamp uint64) *MSRepToken {
	return &MSRepToken{SignedASEntry: signedASEntry, Timestamp: timestamp}
}

func (p *MSRepToken) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *MSRepToken) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *MSRepToken) String() string {
	//TODO (supraja): handle printing the signed as entry better
	return fmt.Sprintf("%s %d", string(p.SignedASEntry), p.Timestamp)
}
