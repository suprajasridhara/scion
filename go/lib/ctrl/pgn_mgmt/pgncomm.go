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

package pgn_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type PGNList struct {
	//L should be a list of SignedPlds in bytes form.
	L         []common.RawBytes
	Timestamp uint64
}

func NewPGNList(l []common.RawBytes, timestamp uint64) *PGNList {
	return &PGNList{L: l, Timestamp: timestamp}
}

func (p *PGNList) ProtoId() proto.ProtoIdType {
	return proto.PGN_TypeID
}

func (p *PGNList) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PGNList) String() string {
	str := ""
	return fmt.Sprintf("%s %d", str, p.Timestamp)
}
