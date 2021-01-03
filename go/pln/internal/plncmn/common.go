// Copyright 2021 ETH Zurich
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
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
)

func Init(cfg plnconfig.PlnConf, sdCfg env.SCIONDClient, features env.Features) error {
	router, err := infraenv.NewRouter(cfg.IA, sdCfg)
	if err != nil {
		return serrors.WrapStr("Error in Init plncmn", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.IA,
		Public:                &net.UDPAddr{IP: cfg.IP, Port: int(cfg.Port)},
		SVC:                   addr.SvcPLN,
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

	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}

	plnmsgr.IA = cfg.IA
	//plncrypto.CfgDir = cfg.CfgDir
	return nil
}
