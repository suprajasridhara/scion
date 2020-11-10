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

#Db to store ms cfg data (default ./ms.db will be created or read from)
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
rpki_entry_balid = "valid" 

#PLNIA IA of the PLN to contact for PCN lists (required)
pln_isd_as = "1-ff00:0:110"

#MSListValidTime time for which a published ms list is valid in minutes (default = 10080) 1 week
ms_list_valid_time = 10080

#MSPullListInterval time intervaal to pull full ms list in minutes (default = 1440) 1 day
ms_pull_list_interval = 1440
`
