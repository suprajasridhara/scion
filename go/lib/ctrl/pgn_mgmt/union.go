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
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

// union represents the contents of the unnamed capnp union.
type union struct {
	Which              proto.PGN_Which
	AddPLNEntryRequest *AddPLNEntryRequest
}

func (u *union) set(c proto.Cerealizable) error {
	switch p := c.(type) {
	case *AddPLNEntryRequest:
		u.Which = proto.PGN_Which_addPLNEntryRequest
		u.AddPLNEntryRequest = p
	default:
		return common.NewBasicError("Unsupported PGN ctrl union type (set)", nil,
			"type", common.TypeOf(c))
	}
	return nil
}

func (u *union) get() (proto.Cerealizable, error) {
	switch u.Which {
	case proto.PGN_Which_addPLNEntryRequest:
		return u.AddPLNEntryRequest, nil
	}
	return nil, common.NewBasicError("Unsupported PLN ctrl union type (get)", nil,
		"type", u.Which)
}
