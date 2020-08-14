// Copyright 2018 Anapaya Systems
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

package config

import (
	"io"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/config"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
)

type Config struct {
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	Sciond   env.SCIONDClient `toml:"sciond_connection,omitempty"`
	Ms       MsConf           `toml:"sig,omitempty"`
}

func (cfg *Config) InitDefaults() {
	config.InitAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Ms,
	)
}

func (cfg *Config) Validate() error {
	return config.ValidateAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Ms,
	)
}

func (cfg *Config) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	config.WriteSample(dst, path, config.CtxMap{config.ID: idSample},
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Ms,
	)
}

var _ config.Config = (*MsConf)(nil)

type MsConf struct {
	//TODO_MS:(supraja) create the config definition to start the MS instance
	//sciond config, IP, PLN config
	ID               string
	DispatcherBypass string
	IP               net.IP
	CtrlPort         uint16
	DataPort         uint16
	IA               addr.IA
}

func (cfg *MsConf) InitDefaults() {
	//TODO_MS:(supraja)

}
func (cfg *MsConf) Validate() error {
	//TODO_MS:(supraja)
	return nil
}

func (cfg *MsConf) ConfigName() string {
	return "ms"
}

func (cfg *MsConf) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	//TODO_MS:(supraja) return config sample
}
