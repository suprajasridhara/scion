package ms_mgmt

import (
	"fmt"
	"strings"
	"time"

	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*Pld)(nil)

type MsgIdType uint64

func (m MsgIdType) String() string {
	return fmt.Sprintf("0x%016x", uint64(m))
}

func (m MsgIdType) Time() time.Time {
	return time.Unix(int64(m)/1000000000, int64(m)%1000000000)
}

//Pld is the generic payload for MS
type Pld struct {
	Id MsgIdType
	union
}

// NewPld creates a new MS ctrl payload, containing the supplied Cerealizable instance.
func NewPld(id MsgIdType, u proto.Cerealizable) (*Pld, error) {
	p := &Pld{Id: id}
	return p, p.union.set(u)
}

func (p *Pld) Union() (proto.Cerealizable, error) {
	return p.union.get()
}

func (p *Pld) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *Pld) String() string {
	desc := []string{fmt.Sprintf("MS: Id: %s Union:", p.Id)}
	u, err := p.Union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u))
	}
	return strings.Join(desc, " ")
}
