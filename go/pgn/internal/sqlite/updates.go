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

const (
	//UpdateNodeListEntry is the query used to update a row in pgn_entries
	// based on srcIA and entryType
	UpdateEntry = `
	Update pgn_entries SET 
	entry = ?,
	commitId = ?,
	timestamp = ?,
	signedBlob = ?
	WHERE srcIA = ? and entryType = ?
	`
)
