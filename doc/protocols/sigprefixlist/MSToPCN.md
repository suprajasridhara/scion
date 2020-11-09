# Requests from MS to PCN

The messages exchanged between MS and PCN can be seen in the figure. 
![ms-sig communication](../../fig/mapping_srv/MS-PCN.svg)

## MS sends mapping list to PCN
The MS sends its signed mapping list to PCNs periodically. The time interval (in minutes) can be specified in the config file during service startups. This time interval should not be more than the global value for validity of MS lists, otherwise the lists in the PCNs will be stale and invalid.

To push the list the MS performs the following:
- get the AS entries in the new_entries table (see [Databases](MappingService.md/#Database))
- fetch the PCN list from the configured PLN (see [MS requests PLN List](MSToPLN.md))
- pick a random PCN from the list
- form the ms_mgmt.SignedMSList payload 
- send the list to the PCN that was picked using the messenger instance stored in msmsgr


