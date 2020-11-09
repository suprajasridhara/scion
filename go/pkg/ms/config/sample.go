package config

const idSample = "ms1"

const msSample = `
# ID of the MS. (required)
id = "%s"

# IP to listen on (required)
IP = "127.0.0.65"
# Port to listen on (required)
Port = 3009 
# IA the local IA (required)
IA = "1-ff00:0:110"

#CfgDir directory to read crypto keys from (required)
CfgDir = "gen/ISD1/ASff00_0_110" 

#Db to store ms cfg data (default ./ms.db will be created or read from)
Db = "./ms.db" 

#QUIC address to listen to quic IP:Port (required)
QUICAddr = "127.0.0.27:20655" 

#CertFile for QUIC socket (required)
CertFile = "gen-certs/tls.pem" 

#KeyFile for QUIC socket (required)
KeyFile = "gen-certs/tls.key" 

#RPKIValidator is the path to the shell scripts that takes 2 arguments,
#ASID and the prefix to validate (required)
RPKIValidator = "validator.sh" 

#RPKIValidString is the response of the validator script if the ASID and prefix are valid (required)
RPKIValidString = "valid" 

#PLNIA IA of the PLN to contact for PCN lists (required)
PLNIA = "1-ff00:0:110"

#MSListValidTime time for which a published ms list is valid in minutes (default = 10080) 1 week
MSListValidTime = 10080

#MSPullListInterval time intervaal to pull full ms list in minutes (default = 1440) 1 day
MSPullListInterval = 1440

`
