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

type AddPGNEntryRequest struct {
	Entry      []byte
	EntryType  string
	CommitID   string
	PGNId      string `capnp:"pgnId"`
	Timestamp  uint64
	SrcIA      string `capnp:"srcIA"`
}

func NewAddPGNEntryRequest(entry []byte, entryType string, commitID string,
	pgnId string, timestamp uint64, srcIA string) *AddPGNEntryRequest {

	return &AddPGNEntryRequest{Entry: entry, EntryType: entryType,
		CommitID: commitID, PGNId: pgnId, Timestamp: timestamp, SrcIA: srcIA}
}

func (p *AddPGNEntryRequest) ProtoId() proto.ProtoIdType {
	return proto.PGN_TypeID
}

func (p *AddPGNEntryRequest) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *AddPGNEntryRequest) String() string {
	return fmt.Sprintf("%s %s %s %s %d %s", p.Entry, p.EntryType, p.CommitID,
		p.PGNId, p.Timestamp, p.SrcIA)
}

type PGNRep struct {
	Entry     AddPGNEntryRequest
	Timestamp uint64
}

func NewPGNRep(entry AddPGNEntryRequest, timestamp uint64) *PGNRep {
	return &PGNRep{Entry: entry, Timestamp: timestamp}
}

func (p *PGNRep) ProtoId() proto.ProtoIdType {
	return proto.PGN_TypeID
}

func (p *PGNRep) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PGNRep) String() string {
	return fmt.Sprintf("%s %d", p.Entry.String(), p.Timestamp)
}

type PGNEntryRequest struct{
	EntryType string
	SrcIA string `capnp:"srcIA"`
}

func NewPGNEntryRequest(entryType string, srcIA string) *PGNEntryRequest {
	return &PGNEntryRequest{EntryType: entryType, SrcIA: srcIA}
}

func (p *PGNEntryRequest) ProtoId() proto.ProtoIdType {
	return proto.PGN_TypeID
}

func (p *PGNEntryRequest) Write(b common.RawBytes) (int, error) {
	return proto.WriteRoot(p, b)
}

func (p *PGNEntryRequest) String() string {
	return fmt.Sprintf("%s %s", p.EntryType, p.SrcIA)
}