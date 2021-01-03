# Publishing Gossip Node

The Publishing Gossip Node (PGN) accepts mapping
lists from [MSes](./MappingService.md) and stores them.
It then uses gossip to propagate these lists to other PGNs
near it. It responds to list queries from MSes.

## Deployment

To deploy a PGN instance run the following command from the
[go/pgn](../../../go/pgn) folder

```sh
go run main.go -config <path_to_config_file>
```

For more information on the configuration, see
[Configuration](#Configuration)

The PGN connects to a [PLN](./PublishingListNode.md) instance.
A PLN instance must be running when starting a PGN instance.
The PGN on startup registers its presence with the PLN.

## Configuration

The sample configuration file can be generated using

```sh
go run main.go -help-config
```

from the [go/pgn](../../../go/pgn).

## General Structure

It reuses existing packages to build up the service;

- [go/lib/env](../../../go/lib/env)--used for configuration and
    setup of the service
- [go/pkg/trust](../../../go/pkg/trust)--used for TRCs and other
    crypto material
- [go/lib/infra](../../../go/pkg/trust)--used for the messenger
    to send and receive messages

### Main folders in PGN

- [go/pgn/internal](../../../go/pgn/internal)--contains functionality
    internal to the PGN and required for its functioning
    - [pgncmn](../../../go/pgn/internal/pcgcmn)--performs actions common
    to all blocks of the PGN. It initializes the network, establishes
    a connection to SCIOND, initializes an instance of the messenger
    and registers handlers. Additionally it saves some state to be
    used by other packages
    - [pgncrypto](../../../go/pgn/internal/pgncrypto)--wrapper
    for [go/pkg/trust](../../../go/pkg/trust)
    - [pgnmsgr](../../../go/pgn/internal/pgnmsgr)--wrapper
    for [go/pkg/infra](../../../go/pkg/infra) and also stores an instance of the messenger
    - [sqlite](../../../go/pgn/internal/sqlite)--handles all database
    operations for the PGN

The other folders are meant for communication with other services in
the protocol. They also contain the handlers for various messages.

## Database

The database is an sqlite3 instance, specified in the configuration
file on starting the service. The service uses the databases
if it exists, otherwise creates one.

The schema of the database is defined in the sqlite3
package's [schema.go](../../../go/pgn/internal/sqlite/schema.go).





