package pcncmn

import (
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pcn/internal/pcncrypto"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/mscomm"
	"github.com/scionproto/scion/go/pcn/pcncomm"
	pcnconfig "github.com/scionproto/scion/go/pkg/pcn/config"
)

var (
	IA       addr.IA
	CtrlAddr net.IP
	CtrlPort int
	PLNIA    addr.IA
)

func Init(cfg pcnconfig.PcnConf, sdCfg env.SCIONDClient, features env.Features) error {
	CtrlAddr = cfg.IP
	CtrlPort = int(cfg.CtrlPort)

	router, err := infraenv.NewRouter(cfg.IA, sdCfg)
	if err != nil {
		return serrors.WrapStr("Error in Init mscmn", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.IA,
		Public:                &net.UDPAddr{IP: cfg.IP, Port: int(cfg.CtrlPort)},
		SVC:                   addr.SvcPCN,
		ReconnectToDispatcher: true,
		QUIC: infraenv.QUIC{
			Address:  cfg.QUICAddr,
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		},
		Router:    router,
		SVCRouter: messenger.NewSVCRouter(itopo.Provider()),
	}
	IA = cfg.IA
	PLNIA = cfg.PLNIA
	pcnmsgr.Msgr, err = nc.Messenger()
	pcnmsgr.IA = cfg.IA
	pcncrypto.CfgDir = cfg.CfgDir

	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}

	//Add messenger handlers here
	pcnmsgr.Msgr.AddHandler(infra.PushMSListRequest, mscomm.MSListHandler{})
	pcnmsgr.Msgr.AddHandler(infra.NodeList, pcncomm.NodeListHandler{})
	pcnmsgr.Msgr.AddHandler(infra.NodeListEntryRequest, mscomm.NodeListEntryReqHandler{})

	return nil
}
