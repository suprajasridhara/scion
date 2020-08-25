@0x8273379c3e06a731;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

using Sciond = import "sciond.capnp";

struct MSMsg {
    id @0 :UInt64;  # Request ID
    union {
        fullMapRec @1 :FullMapRec;
        asIDRec @2 :ASIDReq;
    }
    traceId @3 :Data;
}


struct FullMapRec{
    sigID @0 : UInt64;
}


struct FullMapResp{
         sigID @0 : UInt64;

}

struct ASIDReq{
 sigID @0 : UInt64;
}

struct ASIDRes{
 sigID @0 : UInt64;
}


struct MSAddr {
    ctrl @0 :Sciond.HostInfo;
    data @1 :Sciond.HostInfo;
}
