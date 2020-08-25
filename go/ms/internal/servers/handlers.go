package servers

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/sciond"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn, src net.Addr, pld *ms.Pld)
}

type FullMapRequestHandler struct {
}

func (h *FullMapRequestHandler) Handle(ctx context.Context, conn net.Conn, src net.Addr,
	pld *sciond.Pld) {
}

type ASIDRequestHandler struct {
}

func (h *ASIDRequestHandler) Handle(ctx context.Context, conn net.Conn, src net.Addr,
	pld *sciond.Pld) {
}
