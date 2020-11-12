# Mapping Service
The Mapping Service (MS) acts as a bridge between the Publishing Infrastructure and the SIGs. It is also a trusted entity for the ISD it is deployed in and should be deployed in Core ASs only. 

It is responsible for forming lists of mappings it receives from downstream ASs and submitting them to the Publishing Infrastructure. It pulls mapping lists for other ISDs published in the Publishing Infrastructure and stores it and responds to SIGs that require the mappings. 

## Deployment 
To deploy a Mapping Service instance run the following command from [go/ms](../../../go/ms) folder
```
go run main.go -config <path_to_config_file>
```
For more information on the configuration see 
[Configuration](#Configuration)

The Mapping Service connects to a [PLN](./PublishingListNode.md) instance. The PLN instance must be running when starting a MS instance.

## Configuration
The sample configuration file can be generated using 
```
go run main.go -help-config
```
from [go/ms](../../../go/ms)

## General Structure
It reuses existing packages to build up the service
- [go/lib/env](../../../go/lib/env): Is used for configuration and setup of the service.
- [go/pkg/trust](../../../go/pkg/trust): Is used for TRCs and other crypto material.
- [go/lib/infra](../../../go/pkg/trust) : Is used for the messenger to send and receive messages.

#### Main folders in MS
- [go/ms/internal](../../../go/ms/internal): It contains functionality internal to the Mapping Service and required for its functioning. 
    - [mscmn](../../../go/ms/internal/mscmn): performs actions common to all blocks of the MS. It initializes the network, establishes a connection to SCIOND, initializes an instance of the messenger and registers handlers. Additionally it saves some state to be used by other packages.
    - [mscrypto](../../../go/ms/internal/mscrypto): this is a  wrapper for [go/pkg/trust](../../../go/pkg/trust) 
    - [msmsgr](../../../go/ms/internal/msmsgr):this is a wrapper for [go/pkg/infra](../../../go/pkg/infra) and also stores an instance of the messenger
    - [sqlite3](../../../go/ms/internal/sqlite3): handles all database operations for the MS
    - [validator](../../../go/ms/internal/validator): defines a RPKI validator

The other folders are meant for communication with other services in the protocol. They also contain the handlers for various messages. 

## Database
The database is a sqlite3 instance. To use an existing database the path can be specified in the configuration file during service start up. The service uses the databases if it exists, otherwise it creates one in the specified path.

The schema of the database is defined in the sqlite3 package's [schema.go](../../../go/ms/internal/sqlite3/schema.go)





