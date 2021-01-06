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

const idSample = "pln1"

const plnSample = `
# ID of the PLN. (required)
id = "%s"

# IP to listen on (required)
ip = "127.0.0.65"

# Port to listen on (required)
port = 3009 

# IA the local IA (required)
isd_as = "1-ff00:0:110"

#CfgDir directory to read crypto keys from (required)
cfg_dir = "gen/ISD1/ASff00_0_110"

#Db to store PLN cfg data (default ./pln.db will be created or read from)
db = "./pln.db" 

#QUIC address to listen to quic IP:Port (required)
quic_addr = "127.0.0.27:20655" 

#CertFile for QUIC socket (required)
cert_file = "gen-certs/tls.pem" 

#KeyFile for QUIC socket (required)
key_file = "gen-certs/tls.key" 

#PropagateInterval is the time interval between PLN list propagations (default = 1 hour)
prop_interval = 1
`
