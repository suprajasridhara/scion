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

package ms_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type SignedASEntry struct {
	Blob common.RawBytes
	Sign *proto.SignS
}

func NewSignedASEntry(blob common.RawBytes, sign *proto.SignS) *SignedASEntry {
	return &SignedASEntry{Blob: blob, Sign: sign}
}

func (p *SignedASEntry) ProtoId() proto.ProtoIdType {
	return proto.MS_TypeID
}

func (p *SignedASEntry) String() string {
	return fmt.Sprintf("SignedASEntry: %s %s", p.Blob, p.Sign)
}

type SignedMSList struct {
	Timestamp uint64
	ASEntries []SignedASEntry `capnp:"asEntries"`
	MSIA      string          `capnp:"msIA"`
}

func NewSignedMSList(timestamp uint64, asEntry []SignedASEntry,
	msIA string) *SignedMSList {

	return &SignedMSList{Timestamp: timestamp, ASEntries: asEntry, MSIA: msIA}
}

func (p *SignedMSList) ProtoId() proto.ProtoIdType {
	return proto.SignedMSList_TypeID
}

func (p *SignedMSList) String() string {
	s := ""
	for _, asEntry := range p.ASEntries {
		s = s + asEntry.String()
	}
	return fmt.Sprintf("%d %s %s", p.Timestamp, s, p.MSIA)
}
