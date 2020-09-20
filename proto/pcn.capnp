@0xde42b02816b2b3bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

using Pln = import "pln.capnp";

struct PCN {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        addPLNEntryRequest @2 :AddPLNEntryRequest;
        msListRep @3 :MSListRep;
    }
}

struct AddPLNEntryRequest{
    entry @0 :Pln.PlnListEntry;
}

struct MSListRep{
    signedMSList @0 :Data;
    commitId @1 :Text;
    timestamp @2 :UInt64; #PCN promises to add MSList before timestamp 
}

