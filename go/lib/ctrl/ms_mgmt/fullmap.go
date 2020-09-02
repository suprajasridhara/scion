// Copyright 2017 ETH Zurich
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

package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type FullMap struct {
	Id uint8
	Ip string
	Ia string
}

func NewFullMap(id uint8, ip string, ia string) *FullMap {
	return &FullMap{Id: id, Ip: ip, Ia: ia}
}

func (p *FullMap) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *FullMap) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *FullMap) String() string {
	return fmt.Sprintf("%d %s %s", p.Id, p.Ip, p.Ia)
}

type FullMapReq struct {
	Id uint8 `capnp:"id"`
}

func NewFullMapReq(id uint8) *FullMapReq {
	return &FullMapReq{Id: id}
}

func (p *FullMapReq) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *FullMapReq) String() string {
	return fmt.Sprintf("%d", p.Id)
}

type FullMapRep struct {
	Fm []FullMap `capnp:"fm"`
}

func (p *FullMapRep) String() string {
	var s []string
	for _, fm := range p.Fm {
		s = append(s, fm.String())
	}
	return fmt.Sprintf("%v", s)
}

func NewFullMapRep(fm []FullMap) *FullMapRep {
	return &FullMapRep{Fm: fm}
}

func (p *FullMapRep) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}
