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
// limitations under the License.package config
package config

const idSample = "ms1"

const msSample = `
# ID of the MS. (required)
id = "%s"

# IP to listen on (required)
ip = "127.0.0.65"

# Port to listen on (required)
port = 3009 

# IA the local IA (required)
isd_as = "1-ff00:0:110"

#CfgDir directory to read crypto keys from (required)
cfg_dir = "gen/ISD1/ASff00_0_110" 

#Db to store MS cfg data (default ./ms.db will be created or read from)
db = "./ms.db" 

#QUIC address to listen to quic IP:Port (required)
quic_addr = "127.0.0.27:20655" 

#CertFile for QUIC socket (required)
cert_file = "gen-certs/tls.pem" 

#KeyFile for QUIC socket (required)
key_file = "gen-certs/tls.key" 

#RPKIValidator is the path to the shell scripts that takes 2 arguments,
#ASID and the prefix to validate (required)
rpki_validator = "validator.sh" 

#RPKIValidString is the response of the validator script if the ASID and prefix are valid (required)
rpki_entry_valid = "valid" 

#PLNIA IA of the PLN to contact for PCN lists (required)
pln_isd_as = "1-ff00:0:110"

#MSListValidTime time for which a published MS list is valid in minutes (default = 10080) 1 week
ms_list_valid_time = 10080

#MSPullListInterval time interval to pull full MS list in minutes (default = 1440) 1 day
ms_pull_list_interval = 1440
`
