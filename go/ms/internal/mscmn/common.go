// Copyright 2020 ETH Zurich
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
package mscmn

import (
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/messenger"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/validator"
	"github.com/scionproto/scion/go/ms/sigcomm"
	msconfig "github.com/scionproto/scion/go/pkg/ms/config"
)

var (
	Addr net.IP
	Port int
)

func Init(cfg msconfig.MsConf, sdCfg env.SCIONDClient, features env.Features) error {
	Addr = cfg.IP
	Port = int(cfg.Port)

	router, err := infraenv.NewRouter(cfg.IA, sdCfg)
	if err != nil {
		return serrors.WrapStr("Error in Init mscmn", err)
	}
	nc := infraenv.NetworkConfig{
		IA:                    cfg.IA,
		Public:                &net.UDPAddr{IP: cfg.IP, Port: int(cfg.Port)},
		SVC:                   addr.SvcMS,
		ReconnectToDispatcher: true,
		QUIC: infraenv.QUIC{
			Address:  cfg.QUICAddr,
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		},
		Router:                router,
		SVCRouter:             messenger.NewSVCRouter(itopo.Provider()),
		SVCResolutionFraction: 1, //this ensures that QUIC connection is always used
	}
	msmsgr.Msgr, err = nc.Messenger()
	if err != nil {
		return serrors.WrapStr("Unable to fetch Messenger", err)
	}
	msmsgr.IA = cfg.IA
	mscrypto.CfgDir = cfg.CfgDir
	validator.Path = cfg.RPKIValidator
	validator.EntryValid = cfg.RPKIValidString

	msmsgr.Msgr.AddHandler(infra.MSFullMapRequest, sigcomm.FullMapReqHandler{})
	msmsgr.Msgr.AddHandler(infra.ASActionRequest, sigcomm.ASActionHandler{})

	return nil
}
