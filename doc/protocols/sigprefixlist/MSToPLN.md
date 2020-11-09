# Requests from MS to PLN

The messages exchanged between MS and PLN can be seen in the figure. 
![ms-sig communication](../../fig/mapping_srv/MS-PLN.svg)

## MS requests PLN List
The Mapping Service sends the request using the infra.Messenger instance in msmsgr package and verifies the origin of the response before processing it. It then returns the processed list of PCN Id and IA objects to the calling function 

## Handler in PLN

