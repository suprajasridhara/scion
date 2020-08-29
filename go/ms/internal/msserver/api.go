package msserver

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"net"

// 	"github.com/opentracing/opentracing-go"
// 	opentracingext "github.com/opentracing/opentracing-go/ext"
// 	"github.com/scionproto/scion/go/lib/log"
// 	"github.com/scionproto/scion/go/lib/tracing"
// 	"github.com/scionproto/scion/go/ms/internal/mstypes"
// 	"github.com/scionproto/scion/go/proto"
// 	capnp "zombiezen.com/go/capnproto2"
// )

// // ConnHandler is a MS API server running on top of a PacketConn. It
// // reads messages from the transport, and passes them to the relevant request
// // handler.
// type ConnHandler struct {
// 	Conn net.Conn
// 	// State for request Handlers
// 	Handlers map[proto.MSMsg_Which]Handler
// }

// func (srv *ConnHandler) Serve(address net.Addr) {
// 	msg, err := proto.SafeDecode(capnp.NewDecoder(srv.Conn))
// 	if err != nil {
// 		log.Error("Unable to decode RPC request", "err", err)
// 		return
// 	}

// 	root, err := msg.RootPtr()
// 	if err != nil {
// 		log.Error("Unable to extract capnp root", "err", err)
// 		return
// 	}

// 	p := &mstypes.Pld{}
// 	if err := proto.SafeExtract(p, proto.SCIONDMsg_TypeID, root.Struct()); err != nil {
// 		log.Error("Unable to extract capnp SCIOND payload", "err", err)
// 		return
// 	}

// 	handler, ok := srv.Handlers[p.Which]
// 	if !ok {
// 		log.Error("handler not found for capnp message", "which", p.Which)
// 		return
// 	}

// 	var spanCtx opentracing.SpanContext
// 	if len(p.TraceId) > 0 {
// 		var err error
// 		spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.Binary,
// 			bytes.NewReader(p.TraceId))
// 		if err != nil {
// 			log.Info("Failed to extract span", "err", err)
// 		}
// 	}

// 	span, ctx := tracing.CtxWith(context.Background(), fmt.Sprintf("%s.handler", p.Which),
// 		opentracingext.RPCServerOption(spanCtx))
// 	defer span.Finish()

// 	handler.Handle(ctx, srv.Conn, address, p)
// }

// func (srv *ConnHandler) Close() error {
// 	return srv.Conn.Close()
// }
