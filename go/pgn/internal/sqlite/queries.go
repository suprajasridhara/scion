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
	//EntriesByTypeAndSrcIA queries all rows in pgn_entries by entryType and srcIA
	EntriesByTypeAndSrcIA = "Select * from pgn_entries where entryType like ? and srcIA like ?"

	//EmptyObjects queries all rows in empty_objects
	EmptyObjects = "Select * from empty_objects"
)
