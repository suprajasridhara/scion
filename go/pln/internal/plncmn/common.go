package plncmn

import (
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/serrors"
	plnconfig "github.com/scionproto/scion/go/pkg/pln/config"
	"github.com/scionproto/scion/go/pln/plnmsgr"
)

var (
	IA       addr.IA
	CtrlAddr net.IP
	CtrlPort int
)

func Init(cfg plnconfig.PlnConf, sdCfg env.SCIONDClient, features env.Features) error {
	CtrlAddr = cfg.IP
	CtrlPort = int(cfg.CtrlPort)

	router, err := infraenv.NewRouter(cfg.IA, sdCfg)
	if err != nil {
		return serrors.WrapStr("Error in Init mscmn", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.IA,
		Public:                &net.UDPAddr{IP: cfg.IP, Port: int(cfg.CtrlPort)},
		SVC:                   addr.SvcMS,
		ReconnectToDispatcher: true,
		QUIC: infraenv.QUIC{
			Address:  cfg.QUICAddr,
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		},
		Router:    router,
		SVCRouter: messenger.NewSVCRouter(itopo.Provider()),
	}
	plnmsgr.Msgr, err = nc.Messenger()
	plnmsgr.IA = cfg.IA
	// plncrypto.CfgDir = cfg.CfgDir
	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}

	// plnmsgr.Msgr.AddHandler(infra.MSFullMapRequest, sigcomm.FullMapReqHandler{})
	// plnmsgr.Msgr.AddHandler(infra.ASActionRequest, sigcomm.ASActionHandler{})

	return nil
}