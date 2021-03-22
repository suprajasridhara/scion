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

package sqlite

import "database/sql"

//NodeListEntry sql table struct that to read from sqlite
type PGNEntry struct {
	ID        int            `TbField:"id"`
	Entry     *[]byte        `TbField:"entry"`
	CommitID  sql.NullString `TbField:"commitID"`
	SrcIA     sql.NullString `TbField:"srcIA"`
	Timestamp int            `TbField:"timestamp"`
	EntryType sql.NullString `TbField:"entryType"`
}
