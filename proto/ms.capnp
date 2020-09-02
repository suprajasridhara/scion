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
        asActionReq @4 :ASMapEntry;
        asActionRep @5 :MSRepToken;
    }
}

struct ASMapEntry{
    ia @0 :Text;
    ip @1 :List(Text);
    timestamp @2 :UInt64;
    action @3 :Text;
}

struct MSRepToken{
    asMapEntry @0 :ASMapEntry;
    timestamp @1 :UInt64; #MS promises to add ASMapEntry before timestamp 
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

