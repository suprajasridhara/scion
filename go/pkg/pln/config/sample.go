package config

const idSample = "pln1"

const plnSample = `
#ID of the PLN. (required)
id = "%s"

#IP to listen on (required)
ip = "127.0.0.65"

# Port to listen on (required)
port = 3009 

#IA the local IA (required)
isd_as = "1-ff00:0:110"

#CfgDir directory to read crypto keys from (required)
cfg_dir = "gen/ISD1/ASff00_0_110" 

#Db to store pln cfg data (default ./pln.db will be created or read from)
db = "./pln.db" 

#QUIC address to listen to quic IP:Port (required)
quic_addr = "127.0.0.27:20655" 

#CertFile for QUIC socket (required)
cert_file = "gen-certs/tls.pem" 

#KeyFile for QUIC socket (required)
key_file = "gen-certs/tls.key" 

#PropogateInterval is the time interval between PLN list propogations in minutes (default = 10)
prop_interval = 10
`
