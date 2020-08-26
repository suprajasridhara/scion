package msserver

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/ms/internal/types"
)

//Handler is an interface ... TODO:
type Handler interface {
	Handle(ctx context.Context, conn net.Conn, src net.Addr, pld *types.Pld)
}

type FullMapRequestHandler struct {
}

func (h *FullMapRequestHandler) Handle(ctx context.Context, conn net.Conn, src net.Addr,
	pld *types.Pld) {
}

type ASIDRequestHandler struct {
}

func (h *ASIDRequestHandler) Handle(ctx context.Context, conn net.Conn, src net.Addr,
	pld *types.Pld) {
}
