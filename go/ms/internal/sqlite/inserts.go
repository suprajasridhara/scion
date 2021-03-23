// Copyright 2020 ETH Zurich
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
	//InsNewEntry inserts a row into the new_entries table
	InsNewEntry = `
	INSERT INTO new_entries(entry) VALUES (?)
	`

	//InsFullMapEntry inserts a row into the full_map table
	InsFullMapEntry = `
	INSERT INTO full_map(ip, ia, timestamp) VALUES (?,?,?)
	`

	//InsPGNRep inserts a row into the pgn_reps table
	InsPGNRep = `
	INSERT INTO pgn_reps(fullRep) VALUES (?)
	`
)
