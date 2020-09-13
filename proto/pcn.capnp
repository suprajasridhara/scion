@0xde42b02816b2b3bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

using Pln = import "pln.capnp";

struct PCN {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        dummy @2 :Text;
    }
}

struct AddPLNEntryRequest{
    entry @0 :Pln.PlnListEntry;
}


