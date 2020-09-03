package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

// type SignedASEntry struct {
// 	ASEntry ASEntry
// 	Sign    string
// }

// func NewSignedASEntry(asEntry ASEntry, sign string) *SignedASEntry {
// 	return &SignedASEntry{ASEntry: asEntry, Sign: sign}
// }

// func (p *SignedASEntry) ProtoId() proto.ProtoIdType {
// 	return proto.MS_TypeID
// }

// func (p *SignedASEntry) Write(b common.RawBytes) (int, error) {
// 	return proto.WriteRoot(p, b)
// }

// func (p *SignedASEntry) String() string {
// 	return fmt.Sprintf("%s %s ", p.ASEntry.String(), p.Sign)
// }

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
	ASEntry   ASMapEntry
	Timestamp uint64
}

func NewMSRepToken(asEntry ASMapEntry, timestamp uint64) *MSRepToken {
	return &MSRepToken{ASEntry: asEntry, Timestamp: timestamp}
}

func (p *MSRepToken) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *MSRepToken) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *MSRepToken) String() string {
	return fmt.Sprintf("%s %d", p.ASEntry.String(), p.Timestamp)
}
