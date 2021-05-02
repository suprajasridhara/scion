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
        addPGNEntryRequest @3 :PGNEntry;
        pgnRep @4 :PGNRep;
        pgnList @5 :PGNList;
        pgnEntryRequest @6: PGNEntryRequest;
    }
}

struct AddPLNEntryRequest{
    entry @0 :Pln.PlnListEntry;
}

struct PGNList{
    l @0 :List(Data);
    emptyObjects @1 :List(Data);
    timestamp @2 :UInt64; #timestamp for when the list was transmitted from a PGN
}

struct PGNEntry{
    entry @0 :Data;
    entryType @1:Text;
    commitID @2 :Text;
    pgnId @3 :Text;
    timestamp @4 :UInt64; #timestamp till when the entry is valid for
    srcIA @5 :Text;
}

struct PGNRep{
    entry @0 :PGNEntry;
    timestamp @1 :UInt64; #timestamp for next broadcast in PGN
}

struct PGNEntryRequest{
    entryType @0: Text;
    srcIA @1: Text;
}
