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
	DefaultDb      = "./pgn.db"
	DefaultNumPGNs = 3
)

var (
	DefaultPropagateInterval = duration{1 * time.Hour} //1 hour
	DefaultConnectTimeout    = duration{1 * time.Minute}
)

type Config struct {
	General  env.General `toml:"general,omitempty"`
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	SD       env.SCIONDClient `toml:"sciond_connection,omitempty"`
	PGN      PGNConf          `toml:"pgn,omitempty"`
}

func (cfg *Config) InitDefaults() {
	config.InitAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.SD,
		&cfg.PGN,
	)
}

func (cfg *Config) Validate() error {
	return config.ValidateAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.SD,
		&cfg.PGN,
	)
}

func (cfg *Config) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	print(path)
	config.WriteSample(dst, path, config.CtxMap{config.ID: idSample},
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.SD,
		&cfg.PGN,
	)
}

var _ config.Config = (*PGNConf)(nil)

type PGNConf struct {
	// ID of the PGN (required)
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
	//Db to store PNG cfg data (default ./pgn.db will be created or read from)
	Db string `toml:"db,omitempty"`
	//QUIC address to listen to QUIC IP:Port (required)
	QUICAddr string `toml:"quic_addr,omitempty"`
	//CertFile for QUIC socket (required)
	CertFile string `toml:"cert_file,omitempty"`
	//KeyFile for QUIC socket (required)
	KeyFile string `toml:"key_file,omitempty"`
	//PLNIA IA of the PLN to contact for PGN lists (required)
	PLNIA addr.IA `toml:"pln_isd_as,omitempty"`
	//ConnectTimeout is the amount of time the messenger waits for a reply
	//from the other service that it connects to. default (1 minute)
	ConnectTimeout duration `toml:"connect_timeout,omitempty"`
	//PropagateInterval is the time interval between PGNEntry lists propagations (default = 1 hour)
	PropagateInterval duration `toml:"prop_interval"`
	//NumPGNs is the number of PGNs that the PGNEntry list is propagated to in
	//every interval (default = 3)
	NumPGNs uint16 `toml:"num_pgns"`
	//ISDRange is the range of ISD numbers in the current network (required)
	ISDRange string `toml:"isd_range"`
}

func (cfg *PGNConf) InitDefaults() {
	if cfg.Db == "" {
		cfg.Db = DefaultDb
	}
	if cfg.ConnectTimeout.Duration == 0 {
		cfg.ConnectTimeout = DefaultConnectTimeout
	}
	if cfg.PropagateInterval.Duration == 0 {
		cfg.PropagateInterval = DefaultPropagateInterval
	}
	if cfg.NumPGNs == 0 {
		cfg.NumPGNs = DefaultNumPGNs
	}
}
func (cfg *PGNConf) Validate() error {
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
		return serrors.New("Port must be set")
	}
	if cfg.CfgDir == "" {
		return serrors.New("pgn cfg_dir should be set")
	}
	if cfg.QUICAddr == "" {
		return serrors.New("quic_addr should be set")
	}
	if cfg.CertFile == "" {
		return serrors.New("cert_file must be set")
	}
	if cfg.KeyFile == "" {
		return serrors.New("key_file must be set")
	}
	if cfg.PLNIA.IsZero() {
		return serrors.New("pln_isd_as must be set")
	}
	if cfg.ISDRange == "" {
		return serrors.New("isd_range must be set")
	}
	return nil
}

func (cfg *PGNConf) ConfigName() string {
	return "pgn"
}

func (cfg *PGNConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(pgnSample, ctx[config.ID]))
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
