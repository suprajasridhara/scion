@0xde42b02816bdc1bf;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct MS {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        fullMapReq @2 :FullMapReq;
        fullMapRep @3 :FullMapRep;
    }
}

struct FullMapReq{
    id @0 :UInt8;
}


struct FullMap {
    id @0 :UInt8;
    ip @1 :Text;
    ia @2 :Text;
}

struct FullMapRep{
    fm @0 :List(FullMap);
}

