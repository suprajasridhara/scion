package ms_mgmt

import "github.com/scionproto/scion/go/proto"

type SignedMap struct {
	//TODO: replace string with correct types if there are any
	sign    string
	ipASMap IPASMap
}
type IPASMap struct {
	entries []*Entry
}

type Entry struct {
	//TODO: replace string with correct types if there are any
	IP       string
	RPKISign string
	ASID     string
}

type Pld struct {
	Which   proto.MSMsg_Which
	FullMap *IPASMap
	ASID    *Entry
	TraceId []byte
}
