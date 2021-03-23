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
	/*Schema PGN uses the following tables:
	- pgn_entries to store the pgn entries it recives from services
	 and other PGNs through gossip*/
	Schema = `
	CREATE TABLE IF NOT EXISTS pgn_entries(
		id INTEGER PRIMARY KEY,
		entry BLOB NOT NULL,
		commitID DATA NOT NULL,
		srcIA INTEGER DATA NOT NULL,
		timestamp INTEGER NOT NULL,
		entryType DATA NOT NULL,
		signedBlob BLOB NOT NUll
	);	

	`
)
