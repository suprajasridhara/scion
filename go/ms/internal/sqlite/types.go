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

import "database/sql"

//FullMapRow sql table struct that to store into and fetch from full_map table
type FullMapRow struct {
	ID        int            `TbField:"id"`
	IP        sql.NullString `TbField:"ip"`
	IA        sql.NullString `TbField:"ia"`
	Timestamp int            `TbField:"timestamp"`
}

//NewEntry sql table struct that to store into and fetch from new_entries table
type NewEntry struct {
	entry *[]byte `TbField:"entry"`
}
