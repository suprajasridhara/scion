package pcnmsgr

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/infra"
)

var Msgr infra.Messenger
var IA addr.IA
var Id string