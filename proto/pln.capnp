@0xde42b02816b2c3bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct PLN {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        plnList @2 :PlnList;
        plnListReq @3 :PlnListReq;
    }
}

struct PlnList {
    l @0 :List(PlnListEntry);
}

struct PlnListEntry{
    pgnId @0 :Text;
    ia @1 :UInt64;
    raw @2 :Data;
}

struct PlnListReq{
    action @0 :Text;
}
