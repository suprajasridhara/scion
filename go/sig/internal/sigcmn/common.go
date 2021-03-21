// Copyright 2017 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sigcmn

import (
	"context"
	"net"
	"time"

	"github.com/scionproto/scion/go/dispatcher/dispatcher"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/sig_mgmt"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/pathmgr"
	"github.com/scionproto/scion/go/lib/sciond/fake"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/sock/reliable"
	sigconfig "github.com/scionproto/scion/go/pkg/sig/config"
	"github.com/scionproto/scion/go/sig/internal/snetmigrate"
)

const (
	MaxPort    = (1 << 16) - 1
	SIGHdrSize = 8
)

var (
	DefV4Net = &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, net.IPv4len*8)}
	DefV6Net = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, net.IPv6len*8)}
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
	Msgr       infra.Messenger
	CoreASes   []addr.AS
)

func Init(cfg sigconfig.SigConf, sdCfg env.SCIONDClient, features env.Features) error {
	IA = cfg.IA
	CtrlAddr = cfg.IP
	CtrlPort = int(cfg.CtrlPort)
	DataAddr = cfg.IP
	DataPort = int(cfg.EncapPort)
	network, resolver, err := initNetwork(cfg, sdCfg, features)
	if err != nil {
		return common.NewBasicError("Error creating local SCION Network context", err)
	}
	conn, err := network.Listen(context.Background(), "udp",
		&net.UDPAddr{IP: CtrlAddr, Port: CtrlPort}, addr.SvcSIG)
	if err != nil {
		return common.NewBasicError("Error creating ctrl socket", err)
	}

	CtrlConn = conn
	Network = network
	PathMgr = resolver

	return nil
}

func initNetwork(cfg sigconfig.SigConf,
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

func initNetworkWithFakeSCIOND(cfg sigconfig.SigConf,
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

func initNetworkWithRealSCIOND(cfg sigconfig.SigConf,
	sdCfg env.SCIONDClient, features env.Features) (*snet.SCIONNetwork, pathmgr.Resolver, error) {

	// TODO(karampok). To be kept until https://github.com/scionproto/scion/issues/3377
	deadline := time.Now().Add(sdCfg.InitialConnectPeriod.Duration)
	var retErr error
	for tries := 0; time.Now().Before(deadline); tries++ {
		resolver, err := snetmigrate.ResolverFromSD(sdCfg.Address, sdCfg.PathCount)
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
		log.Debug("SIG is retrying to get NewNetwork", "err", err)
		retErr = err
		time.Sleep(time.Second)
	}
	return nil, nil, retErr
}

func newDispatcher(cfg sigconfig.SigConf) (reliable.Dispatcher, error) {
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

func ValidatePort(desc string, port int) error {
	if port < 1 || port > MaxPort {
		return common.NewBasicError("Invalid port", nil,
			"min", 1, "max", MaxPort, "actual", port, "desc", desc)
	}
	return nil
}

func GetMgmtAddr() sig_mgmt.Addr {
	return *sig_mgmt.NewAddr(addr.HostFromIP(CtrlAddr), uint16(CtrlPort),
		addr.HostFromIP(DataAddr), uint16(DataPort))
}

//SetupMessenger initializes a messenger to be used for SIG-MS communications
func SetupMessenger(cfg sigconfig.Config) error {
	log.Info("Starting messenger initialization")
	router, err := infraenv.NewRouter(cfg.Sig.IA, cfg.Sciond)
	if err != nil {
		return serrors.WrapStr("Unable to fetch router", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.Sig.IA,
		Public:                &net.UDPAddr{IP: cfg.Sig.IP, Port: int(cfg.Sig.UDPPort)},
		SVC:                   addr.SvcSIG,
		ReconnectToDispatcher: false,
		QUIC: infraenv.QUIC{
			Address:  cfg.Sig.QUICAddr,
			CertFile: cfg.Sig.CertFile,
			KeyFile:  cfg.Sig.KeyFile,
		},
		Router:                router,
		SVCRouter:             messenger.NewSVCRouter(itopo.Provider()),
		SVCResolutionFraction: 1, //this ensures that QUIC connection is always used
		ConnectTimeout:        cfg.Sig.MSConnectTimeout,
	}
	Msgr, err = nc.Messenger()
	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}
	log.Info("Messenger initialized successfully")

	return nil
}
