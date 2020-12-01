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
	/*Schema for MS:
	- full_map - stores the processed mapping entries that the MS
		pulls from the Publishing Infrastructure
	- new_entries - stores new mappings that MS recives from SIGs
		to be pushed to the Publishing Infrastructure
	- pcn_reps - stores response tokens from the publishing
		infrastructure.
	*/
	Schema = `
	CREATE TABLE IF NOT EXISTS full_map(
		id INTEGER NOT NULL,
		ip DATA NOT NULL,
		ia DATA NOT NULL,
		timestamp INTEGER NOT NULL,
		PRIMARY KEY (id)
	);

	CREATE TABLE IF NOT EXISTS new_entries(
		id INTEGER PRIMARY KEY,
		entry BLOB
	);

	CREATE TABLE IF NOT EXISTS pcn_reps(
		id INTEGER PRIMARY KEY,
		fullRep BLOB
	)
	`
)
