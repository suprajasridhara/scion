// Copyright 2018 ETH Zurich
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
	//FullMap queries for all rows in the full_map table
	FullMap = `
	SELECT * FROM full_map
	`

	//FullMapEntryByIP queries for all rows in the full_map table that match the ip
	FullMapEntryByIP = `
	SELECT * from full_map where ip = ?
	`

	//NewEntryById queries for all rows in the new_entries table that match the id
	NewEntryById = `
	SELECT entry FROM new_entries where id = ?
	`

	//NewEntries queries for entries in the new_entries table
	NewEntries = `
	SELECT entry FROM new_entries
	`
)
