// Copyright 2021 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pln_mgmt

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
	return time.Unix(int64(m)/1e9, int64(m)%1e9)
}

type Pld struct {
	Id MsgIdType
	union
}

// NewPld creates a new PLN ctrl payload, containing the supplied Cerealizable instance.
func NewPld(id MsgIdType, u proto.Cerealizable) (*Pld, error) {
	p := &Pld{Id: id}
	return p, p.union.set(u)
}

func (p *Pld) Union() (proto.Cerealizable, error) {
	return p.union.get()
}

func (p *Pld) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *Pld) String() string {
	desc := []string{fmt.Sprintf("PLN: Id: %s Union:", p.Id)}
	u, err := p.Union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u))
	}
	return strings.Join(desc, " ")
}
