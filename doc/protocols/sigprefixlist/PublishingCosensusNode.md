# Publishing Consensus Node
The Publishing Consensus Node (PCN) accepts mapping lists from [MSs](./MappingService.md) and stores them. It then uses gossip to propogate these lists to other PCNs near it. It responds to list queries from MSs. 

## Deployment 
To deploy a Mapping Service instance run the following command from [go/pcn](../../../go/pcn) folder
```
go run main.go -config <path_to_config_file>
```
For more information on the configuration see 
[Configuration](#Configuration)

The PCN connects to a [PLN](./PublishingListNode.md) instance. The PLN instance must be running when starting a PCN instance. The PCN on startup registers its presence with the PLN. 

## Configuration
The sample configuration file can be generated using 
```
go run main.go -help-config
```
from [go/pcn](../../../go/pcn)

## General Structure
It reuses existing packages to build up the service
- [go/lib/env](../../../go/lib/env): Is used for configuration and setup of the service.
- [go/pkg/trust](../../../go/pkg/trust): Is used for TRCs and other crypto material.
- [go/lib/infra](../../../go/pkg/trust) : Is used for the messenger to send and receive messages.

### Main folders in PCN
- [go/pcn/internal](../../../go/pcn/internal): It contains functionality internal to the Mapping Service and required for its functioning. 
    - [pcncmn](../../../go/pcn/internal/mscmn): performs actions common to all blocks of the mapping service. It initializes the network, establishes a connection to SCIOND, initializes an instance of the messenger and registers handlers. Additionally it saves some state to be used by other packages.
    - [pcncrypto](../../../go/pcn/internal/mscrypto): this is a  wrapper for [go/pkg/trust](../../../go/pkg/trust) 
    - [pcnmsgr](../../../go/pcn/internal/msmsgr):this is a wrapper for [go/pkg/infra](../../../go/pkg/infra) and also stores an instance of the messenger
    - [sqlite](../../../go/pcn/internal/sqlite): handles all database operations for the mapping service

The other folders are meant for communication with other services in the protocol. They also contain the handlers for various messages. 

## Database
The database is a sqlite3 instance, specified in the configuration file on starting the service. The service uses the databases if it exists, otherwise creates one.

The schema of the database is defined in the sqlite3 package's [schema.go](../../../go/pcn/internal/sqlite/schema.go)





