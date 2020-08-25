package msserver

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/proto"
)

type ServerCfg struct {
}

type HandlerMap map[proto.MSMsg_Which]Handler

type Server struct {
	network  string
	address  string
	handlers map[proto.MSMsg_Which]Handler

	mu          sync.Mutex
	listener    net.Listener
	closeCalled bool
}

// Server constructs a API server. The caller is responsible for starting and
// shutting it down.
func HandleAndServe(addr string, cfg ServerCfg) *Server {
	handlers := servers.HandlerMap{
		proto.MSMsg_Which_fullMapRec: &FullMapRequestHandler{},
		proto.MSMsg_Which_asIDRec:    &ASIDRequestHandler{},
	}
	return NewServer("tcp", addr, handlers)
}

// NewServer initializes a new server at address on the specified network. The
// server will route requests to their correct handlers based on the
// HandlerMap. To start listening on the address, call ListenAndServe.
//
// Network must be "unixpacket" or "rsock".
func NewServer(network string, address string, handlers HandlerMap) *Server {
	return &Server{
		network:  network,
		address:  address,
		handlers: handlers,
	}
}

// ListenAndServe starts listening on srv's address, and repeatedly accepts
// connections from clients. For each accepted connection, a SCIONDMsg server
// is started as a separate goroutine; the server will manage the connection
// until it is closed by the client.
func (srv *Server) ListenAndServe() error {
	srv.mu.Lock()
	if srv.closeCalled {
		srv.mu.Unlock()
		return serrors.New("attempted to listen on server that was shut down")
	}
	listener, err := net.Listen(srv.network, srv.address)
	if err != nil {
		srv.mu.Unlock()
		return common.NewBasicError("unable to listen on socket", nil,
			"address", srv.address, "err", err)
	}
	srv.listener = listener
	srv.mu.Unlock()
	log.Info("Host API Server started listening", "address", listener.Addr())

	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return err
			}
			log.Info("unable to accept conn", "err", err)
			continue
		}

		go func() {
			defer log.HandlePanic()
			hdl := &ConnHandler{
				Conn:     conn,
				Handlers: srv.handlers,
			}
			hdl.Serve(conn.RemoteAddr())
		}()
	}
}

// Close makes the Server stop listening for new connections, and immediately
// closes all running SCIONDMsg servers that have been launched by this server.
func (srv *Server) Close() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.listener == nil {
		return serrors.New("uninitialized server")
	}
	srv.closeCalled = true
	return srv.listener.Close()
	// FIXME(scrye): shut down running servers once we actually implement the
	// handlers.
}

// Shutdown makes the Server stop listening for new connections, and cleanly
// shuts down all running SCIONDMsg servers that have been launched by this
// server.
func (srv *Server) Shutdown(ctx context.Context) error {
	// Ignore context during close as it should rarely block for non-negligible
	// time.
	if err := srv.Close(); err != nil {
		return err
	}

	// FIXME(scrye): cleanly close running SCIONDMsg servers here.
	return nil
}
