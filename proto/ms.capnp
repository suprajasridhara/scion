@0xde42b02816bdc1bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct MS {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        fullMapReq @2 :FullMap;
        fullMapRep @3 :FullMap;
    }
}


struct FullMap {
    addr @0 :UInt8;
}

