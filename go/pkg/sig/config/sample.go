// Copyright 2018 Anapaya Systems
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

const idSample = "sig4"

const sigSample = `
# ID of the SIG. (required)
id = "%s"

# The SIG config json file. (required)
sig_config = "/etc/scion/sig/sig.json"

# The local ISD-AS. (required)
isd_as = "1-ff00:0:113"

# The bind IP address. (required)
ip = "192.0.2.100"

# Control data port, e.g. keepalives. (default 30256)
ctrl_port = 30256

# Encapsulation data port. (default 30056)
encap_port = 30056

# Name of TUN device to create. (default DefaultTunName)
tun = "sig"

# Id of the routing table. (default 11)
tun_routing_table_id = 11

#Config directory to read crypto material from 
cfg_dir = "etc/gen/ISD1/ASff00_0_113"

#DB file to read/store sig configuration data (default ./sig.db)
db = ./sig.db

#UDP port to open a messenger connection on
udp_port = 30955

#QUIC IP:Port
quic_addr = "192.0.2.111:20655"

#CertFile for QUIC socket
cert_file = "/etc/gen-certs/tls.pem"

#KeyFile for QUIC socket
key_file = "/etc/gen-certs/tls.key"

#PrefixFile contains the list of prefixes that should be pushed to a 
#Mapping service in the ISD. This file is scanned periodically for changes
prefix_file = "/etc/scion/sig/prefixes.json"

#PrefixPushInterval in minutes is the interval between 2 consecutive 
#pushes of prefixes to the mapping service. default (60)
prefix_push_interval = 60 


`
