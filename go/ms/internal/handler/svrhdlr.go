package server

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"

	"github.com/lucas-clemente/quic-go"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/integration"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"github.com/scionproto/scion/go/ms/internal/mscmn"
)

const errorNoError quic.ErrorCode = 0x100

type MsgType uint16

type Server struct {
	quicServer *rpc.Server
}

func (*Server) Init() {
	//TODO: init handler
}

type message struct {
	MsgType MsgType
	Content string
}

const (
	SigPull MsgType = 0x0001
)

func (s Server) Listen() {
	qsock, err := squic.Listen(mscmn.Network, &net.UDPAddr{IP: mscmn.CtrlAddr, Port: mscmn.CtrlPort}, addr.SvcMS, nil)
	if err != nil {
		log.Error("Unable to listen", "err", err)
	}
	if len(os.Getenv(integration.GoIntegrationEnv)) > 0 {
		// Needed for integration test ready signal.
		fmt.Printf("Port=%d\n", qsock.Addr().(*net.UDPAddr).Port)
		fmt.Printf("%s%s\n", integration.ReadySignal, mscmn.IA)
	}
	log.Info("Listening", "ms", qsock.Addr())
	for {
		qsess, err := qsock.Accept(context.Background())
		if err != nil {
			log.Error("Unable to accept quic session", "err", err)
			// Accept failing means the socket is unusable.
			break
		}
		log.Info("Quic session accepted", "src", qsess.RemoteAddr())
		go func() {
			defer log.HandlePanic()
			s.handleClient(qsess)
		}()
	}
}

func (s Server) handleClient(qsess quic.Session) {
	defer qsess.CloseWithError(errorNoError, "")
	qstream, err := qsess.AcceptStream(context.Background())
	if err != nil {
		log.Error("Unable to accept quic stream", "err", err)
		return
	}
	defer qstream.Close()

	qs := newQuicStream(qstream)
	for {
		// Receive ping message
		msg, err := qs.ReadMsg()
		if err != nil {
			if err == io.EOF {
				log.Info("Quic session ended", "src", qsess.RemoteAddr())
			} else {
				log.Error("Unable to read", "err", err)
			}
			break
		}

		switch {
		case msg.MsgType == SigPull:
			//TODO_MS: (supraja) handle sig pull request here
		}

	}
}

type quicStream struct {
	qstream quic.Stream
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func newQuicStream(qstream quic.Stream) *quicStream {
	return &quicStream{
		qstream,
		gob.NewEncoder(qstream),
		gob.NewDecoder(qstream),
	}
}

func (qs quicStream) ReadMsg() (*message, error) {
	var msg message
	err := qs.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, err
}
