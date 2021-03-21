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

type PlnListReq struct {
	Action string
}

func NewPlnListReq(action string) *PlnListReq {
	return &PlnListReq{Action: action}
}

func (p *PlnListReq) ProtoId() proto.ProtoIdType {
	return proto.PLN_TypeID
}

func (p *PlnListReq) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PlnListReq) String() string {
	return fmt.Sprintf("%s ", p.Action)
}
