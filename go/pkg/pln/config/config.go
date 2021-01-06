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

package config

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/config"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
)

const (
	DefaultDb                = "./pln.db"
	DefaultPropagateInterval = 1 * time.Hour //1 hour
)

type Config struct {
	General  env.General `toml:"general,omitempty"`
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	Sciond   env.SCIONDClient `toml:"sciond_connection,omitempty"`
	Pln      PlnConf          `toml:"pln,omitempty"`
}

func (cfg *Config) InitDefaults() {
	config.InitAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pln,
	)
}

func (cfg *Config) Validate() error {
	return config.ValidateAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pln,
	)
}

func (cfg *Config) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	print(path)
	config.WriteSample(dst, path, config.CtxMap{config.ID: idSample},
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pln,
	)
}

var _ config.Config = (*PlnConf)(nil)

type PlnConf struct {
	// ID of the PLN  (required)
	ID string `toml:"id,omitempty"`

	// DispatcherBypass is the underlay address (e.g. ":30041") to use when bypassing SCION
	// dispatcher. If the field is empty bypass is not done and SCION dispatcher is used
	// instead.
	DispatcherBypass string `toml:"disaptcher_bypass,omitempty"`
	// IP to listen on (required)
	IP net.IP `toml:"ip,omitempty"`
	// Port to listen on (required)
	Port uint16 `toml:"port,omitempty"`
	// IA the local IA (required)
	IA addr.IA `toml:"isd_as,omitempty"`

	//CfgDir directory to read crypto keys from (required)
	CfgDir string `toml:"cfg_dir,omitempty"`

	//Db to store PLN cfg data (default ./pln.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//QUIC address to listen to QUIC IP:Port (required)
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket (required)
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket (required)
	KeyFile string `toml:"key_file,omitempty"`

	//PropagateInterval is the time interval between PLN list propagations (default = 1 hour)
	PropagateInterval time.Duration `toml:"prop_interval"`
}

func (cfg *PlnConf) InitDefaults() {
	if cfg.Db == "" {
		cfg.Db = DefaultDb
	}
	if cfg.PropagateInterval == 0 {
		cfg.PropagateInterval = DefaultPropagateInterval
	}
}
func (cfg *PlnConf) Validate() error {

	if cfg.ID == "" {
		return serrors.New("id must be set!")
	}
	if cfg.IA.IsZero() {
		return serrors.New("isd_as must be set")
	}
	if cfg.IA.IsWildcard() {
		return serrors.New("Wildcard isd_as not allowed")
	}
	if cfg.IP.IsUnspecified() {
		return serrors.New("ip must be set")
	}
	if cfg.Port == 0 {
		return serrors.New("port must be set")
	}
	if cfg.CfgDir == "" {
		return serrors.New("MS cfg_dir should be set")
	}
	if cfg.QUICAddr == "" {
		return serrors.New("QUIC addr should be set")
	}
	if cfg.CertFile == "" {
		return serrors.New("cert_file must be set")
	}
	if cfg.KeyFile == "" {
		return serrors.New("key_file must be set")
	}

	return nil
}

func (cfg *PlnConf) ConfigName() string {
	return "pln"
}

func (cfg *PlnConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(plnSample, ctx[config.ID]))
}
