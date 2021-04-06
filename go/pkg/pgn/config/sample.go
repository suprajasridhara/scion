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

package config

const idSample = "pgn1"

const pgnSample = `
# ID of the PGN. (required)
id = "%s"
# IP to listen on (required)
ip = "127.0.0.65"
# Port to listen on (required)
port = 3011 
# IA the local IA (required)
isd_as = "1-ff00:0:110"
#CfgDir directory to read crypto keys from (required)
cfg_dir = "gen/ISD1/ASff00_0_110" 
#Db to store pgn cfg data (default ./pgn.db will be created or read from)
db = "./pgn.db" 
#QUIC address to listen to quic IP:Port (required)
quic_addr = "127.0.0.27:20655" 
#CertFile for QUIC socket (required)
cert_file = "gen-certs/tls.pem" 
#KeyFile for QUIC socket (required)
key_file = "gen-certs/tls.key" 
#PLNIA IA of the PLN to contact for PGN lists (required)
pln_isd_as = "1-ff00:0:110"
#ConnectTimeout is the amount of time the messenger waits for a reply
#from the other service that it connects to. default (1 minute)
connect_timeout = "1m"
#PropagateInterval is the time interval between PGNEntry list propagations (default = 1 hour)
prop_interval = "1h"
#NumPGNs is the number of PGNs that the PGNEntry list is propagated 
#to in every interval (default = 3)
num_pgns = 3
`
