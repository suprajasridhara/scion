@0xde42b02816b2b3bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

using Pln = import "pln.capnp";

struct PGN {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        addPLNEntryRequest @2 :AddPLNEntryRequest;
        nodeList @3 :NodeList;
    }
}

struct AddPLNEntryRequest{
    entry @0 :Pln.PlnListEntry;
}

struct NodeList{
    l @0 :List(NodeListEntry);
    timestamp @1 :UInt64;
}

struct NodeListEntry{
    signedMSList @0 :Data;
    commitId @1 :Text;
}