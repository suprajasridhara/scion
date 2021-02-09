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

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type PlnListEntry struct {
	PGNId string `capnp:"pgnId"`
	IA    uint64 `capnp:"ia"`
	Raw   []byte `capnp:"raw"`
}

func NewPlnListEntry(pgnId string, ia uint64, raw []byte) *PlnListEntry {
	return &PlnListEntry{PGNId: pgnId, IA: ia, Raw: raw}
}

func (p *PlnListEntry) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnListEntry) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnListEntry) String() string {
	return fmt.Sprintf("%s %d %s", p.PGNId, p.IA, p.Raw)
}

type PlnList struct {
	L []PlnListEntry `capnp:"l"`
}

func NewPlnList(l []PlnListEntry) *PlnList {
	return &PlnList{L: l}
}

func (p *PlnList) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnList) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnList) String() string {
	var s []string
	for _, l := range p.L {
		s = append(s, l.String())
	}
	return fmt.Sprintf("%v", s)
}
