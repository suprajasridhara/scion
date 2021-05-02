@0xa1a908676d0a7d2e;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

using Sign = import "sign.capnp";

struct MS {
    id @0 :UInt64;  # Request ID
    union {
        unset @1 :Void;
        asActionReq @2 :ASMapEntry;
        asActionRep @3 :MSRepToken;
        fullMapReq @4 :FullMapReq;
        fullMapRep @5 :FullMapRep;
        signedMSList @6 :SignedMSList;
    }
}

struct ASMapEntry{
    ia @0 :Text;
    ip @1 :List(Text);
    timestamp @2 :UInt64;
    action @3 :Text;
}

struct MSRepToken{
    signedASEntry @0 :Data;
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

struct SignedMSList{
    timestamp @0 :UInt64;
    asEntries @1 :List(SignedEntry); #the payload is of type ASMapEntry with signature from the source AS
    msIA @2 :Text;
}

struct SignedEntry {
    blob @0 :Data; 
    sign @1 :Sign.Sign;
}



