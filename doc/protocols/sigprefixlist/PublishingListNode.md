# Publishing List Node
The Publishing List Node (PLN) is responsible for discovering [PCNs](./PublishingConsensusNode.md). It uses a gossip protocol with other PLN instances near it and accumulates a list of PCNs over time. PCNs register their presence with a PLN they are configured with on start up. [MSs](./MappingService.md) and PCNs pull these lists when they need to find a PCN to communicate with.

## Deployment 
To deploy a PLN instance run the following command from [go/pln](../../../go/pln) folder
```
go run main.go -config <path_to_config_file>
```
For more information on the configuration see 
[Configuration](#Configuration)

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

### Main folders in PLN
- [go/pcn/internal](../../../go/pln/internal): It contains functionality internal to the PLN and required for its functioning. 
    - [plncmn](../../../go/pln/internal/plncmn): performs actions common to all blocks of the PLN. It initializes the network, establishes a connection to SCIOND, initializes an instance of the messenger and registers handlers. Additionally it saves some state to be used by other packages.
    - [plncrypto](../../../go/pln/internal/plncrypto): this is a  wrapper for [go/pkg/trust](../../../go/pkg/trust) 
    - [plnmsgr](../../../go/pln/internal/plnmsgr):this is a wrapper for [go/pkg/infra](../../../go/pkg/infra) and also stores an instance of the messenger
    - [sqlite](../../../go/pln/internal/sqlite): handles all database operations for the PLN

The other folders are meant for communication with other services in the protocol. They also contain the handlers for various messages. 

## Database
The database is a sqlite3 instance, specified in the configuration file on starting the service. The service uses the databases if it exists, otherwise creates one.

The schema of the database is defined in the sqlite package's [schema.go](../../../go/pln/internal/sqlite/schema.go)





