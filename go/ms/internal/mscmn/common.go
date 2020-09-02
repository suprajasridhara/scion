package mscmn

import (
	"context"
	"net"
	"time"

	"github.com/scionproto/scion/go/dispatcher/dispatcher"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/sciond/fake"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/sock/reliable"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/sigreq"
	msconfig "github.com/scionproto/scion/go/pkg/ms/config"
)

var (
	IA         addr.IA
	PathMgr    pathmgr.Resolver
	Dispatcher reliable.Dispatcher
	Network    *snet.SCIONNetwork
	CtrlAddr   net.IP
	CtrlPort   int
	DataAddr   net.IP
	DataPort   int
	CtrlConn   *snet.Conn
)

func Init(cfg msconfig.MsConf, sdCfg env.SCIONDClient, features env.Features) error {
	CtrlAddr = cfg.IP
	CtrlPort = int(cfg.CtrlPort)

	router, err := infraenv.NewRouter(cfg.IA, sdCfg)
	if err != nil {
		return serrors.WrapStr("Unable to fetch router", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.IA,
		Public:                &net.UDPAddr{IP: cfg.IP, Port: int(cfg.CtrlPort)},
		SVC:                   addr.SvcMS,
		ReconnectToDispatcher: true, //TODO (supraja): see later
		QUIC: infraenv.QUIC{
			//TODO (supraja): read all of this from config
			Address:  "127.0.0.12:30755",
			CertFile: "/home/ssridhara/go/src/github.com/scionproto/scion/gen-certs/tls.pem",
			KeyFile:  "/home/ssridhara/go/src/github.com/scionproto/scion/gen-certs/tls.key",
		},
		Router:    router,
		SVCRouter: messenger.NewSVCRouter(itopo.Provider()),
	}

	msmsgr.Msgr, err = nc.Messenger()
	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}

	msmsgr.Msgr.AddHandler(infra.MSFullMapRequest, sigreq.FullMapReqHandler{})
	msmsgr.Msgr.AddHandler(infra.ASActionRequest, sigreq.ASActionHandler{})

	return nil
}

func initNetwork(cfg msconfig.MsConf,
	sdCfg env.SCIONDClient, features env.Features) (*snet.SCIONNetwork, pathmgr.Resolver, error) {

	var err error
	Dispatcher, err = newDispatcher(cfg)
	if err != nil {
		return nil, nil, serrors.WrapStr("unable to initialize SCION dispatcher", err)
	}
	if sdCfg.FakeData != "" {
		return initNetworkWithFakeSCIOND(cfg, sdCfg, features)
	}
	return initNetworkWithRealSCIOND(cfg, sdCfg, features)
}

func newDispatcher(cfg msconfig.MsConf) (reliable.Dispatcher, error) {
	if cfg.DispatcherBypass == "" {
		log.Info("Regular SCION dispatcher", "addr", cfg.DispatcherBypass)
		return reliable.NewDispatcher(""), nil
	}
	// Initialize dispatcher bypass.
	log.Info("Bypassing SCION dispatcher", "addr", cfg.DispatcherBypass)
	dispServer, err := dispatcher.NewServer(cfg.DispatcherBypass, nil, nil)
	if err != nil {
		return nil, serrors.WrapStr("unable to initialize bypass dispatcher", err)
	}
	go func() {
		defer log.HandlePanic()
		err := dispServer.Serve()
		if err != nil {
			log.Error("Bypass dispatcher failed", "err", err)
		}
	}()
	return dispServer, nil
}

func initNetworkWithFakeSCIOND(cfg msconfig.MsConf,
	sdCfg env.SCIONDClient, features env.Features) (*snet.SCIONNetwork, pathmgr.Resolver, error) {

	sciondConn, err := fake.NewFromFile(sdCfg.FakeData)
	if err != nil {
		return nil, nil, serrors.WrapStr("unable to initialize fake SCIOND service", err)
	}
	pathResolver := pathmgr.New(sciondConn, pathmgr.Timers{}, sdCfg.PathCount)
	network := &snet.SCIONNetwork{
		LocalIA: cfg.IA,
		Dispatcher: &snet.DefaultPacketDispatcherService{
			Dispatcher:  Dispatcher,
			SCMPHandler: snet.NewSCMPHandler(pathResolver),
			Version2:    features.HeaderV2,
		},
	}
	return network, pathResolver, nil
}

func initNetworkWithRealSCIOND(cfg msconfig.MsConf,
	sdCfg env.SCIONDClient, features env.Features) (*snet.SCIONNetwork, pathmgr.Resolver, error) {

	// TODO(karampok). To be kept until https://github.com/scionproto/scion/issues/3377
	// TODO:from sig code, keep this?
	deadline := time.Now().Add(sdCfg.InitialConnectPeriod.Duration)
	var retErr error
	for tries := 0; time.Now().Before(deadline); tries++ {
		print(sdCfg.Address)
		resolver, err := ResolverFromSD(sdCfg.Address, sdCfg.PathCount)

		if err == nil {
			return &snet.SCIONNetwork{
				LocalIA: cfg.IA,
				Dispatcher: &snet.DefaultPacketDispatcherService{
					Dispatcher:  Dispatcher,
					SCMPHandler: snet.NewSCMPHandler(resolver),
					Version2:    features.HeaderV2,
				},
			}, resolver, nil
		}
		log.Debug("MS is retrying to get NewNetwork", "err", err)
		retErr = err
		time.Sleep(time.Second)
	}
	return nil, nil, retErr
}

func ResolverFromSD(sciondPath string, pathCount uint16) (pathmgr.Resolver, error) {
	//TODO: see why I need to do this snetmigrate?
	var pathResolver pathmgr.Resolver
	if sciondPath != "" {
		sciondConn, err := sciond.NewService(sciondPath).Connect(
			context.Background())
		if err != nil {
			return nil, serrors.WrapStr("Unable to initialize SCIOND service", err)
		}
		pathResolver = pathmgr.New(sciondConn, pathmgr.Timers{}, pathCount)
	}

	return pathResolver, nil
}
