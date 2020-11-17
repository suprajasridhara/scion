# SIG Prefix List Protocol

## Overview

The protocol is designed to enable SCION-IP gateways (SIGs) to automatically
fetch IP prefixes--IA mappings. It uses a distributed coordinated
infrastructure with the following design and security goals.

The system should

- prevent hijacking attacks, i.e., an AS should not be
    able to add mappings for IP prefixes that it does
    not own;
- prevent flooding attacks on SIGs, i.e., a SIG should
    only be able to create mappings to itself;
- be resilient to downgrade attacks to IP Internet where
    SCION connections are possible; and
- ensure high availability, i.e., the system should not
    have components that can be single points of failure.

To achieve the above properties the following services are used

- Mapping Service ([MS](./MappingService.md))
- SCION-IP Gateway (SIG)
- Publishing Infrastructure Services
    - Publishing List Node ([PLN](./PublishingListNode.md))
    - Publishing Consensus Node ([PCN](./PublishingConsensusNode.md))

## Services

### SIG

The SIG can perform the following actions:

- Get mapping from MS:
    - A SIG that requires a mapping from an IP to an IA queries the MS in Core ASes.
- Add mapping
    - To add a mapping for the AS that the SIG is deployed in, it submits the mapping
    to an MS in the Core ASes.

### Mapping Service (MS)

An MS should be deployed in at least one Core AS of an ISD that supports the use of SIGs.

The MSes performs the following actions:

- Submit lists of mappings to Publishing Infrastructure
    - The MS aggregates entries from different SIGs to submit to the Publishing Infrastructure.
- Reply to SIG mapping queries
    - The MS responds to SIG mapping queries with IP prefix--IA mappings
- Pull lists of mappings from Publishing Infrastructure.
    - The MS pulls lists of mappings from the Publishing Infrastructure and stores
    it to be used to respond to SIG queries.
- Reply to Publishing Infrastructure queries of mapping lists for the ISD

### Publishing Infrastructure Services

#### Publishing List Node (PLN)

This service allows MSes and PCNs to discover PCN locations on the network by maintaining
a list of PCNs that it has discovered through gossip.

It performs the following actions:

- Accept PCN entries from PCNs and store it
- Periodically broadcast the list of PCNs it has discovered to other PLNs
- Accept broadcast list from other PLNs and update its list
- Reply to PCN list queries from MSes and PCNs

#### Publishing Consensus Node (PCN)

This service stores lists with mappings that it receives from MSes and lists that
it receives from other PCNs through gossip.

It performs the following actions:

- Accepts mapping lists from MS and store it
- Periodically broadcast mapping lists it has stored to other PCNs
- Accept broadcast messages from other PCNs with mapping lists and update/add to its lists
- Query MS in an ISD for the mapping list of the ISD

## Security

To prevent hijacking attacks the protocol uses RPKI trust anchors to
validate IP prefix--IA mappings and ascertain ownership of IP prefixes.

To prevent flooding attacks against SIGs it is essential to enforce
that an AS can create mappings only for itself. For this, the protocol
uses the SCION control plane PKI.

To prevent downgrade attacks, a mapping entry in the system must always
be returned and empty responses should be authenticated by a minimum
number of PCNs.