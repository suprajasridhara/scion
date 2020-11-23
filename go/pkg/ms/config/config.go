// Copyright 2018 ETH Zurich
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
	DefaultDb                 = "./ms.db"
	DefaultMSListValidTime    = 10080
	DefaultMSPullListInterval = 1440
)

type Config struct {
	General  env.General `toml:"general,omitempty"`
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	Sciond   env.SCIONDClient `toml:"sciond_connection,omitempty"`
	Ms       MsConf           `toml:"ms,omitempty"`
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
	print(path)
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
	// ID of the MS (required)
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

	//Db to store MS cfg data (default ./ms.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//QUIC address to listen to QUIC IP:Port (required)
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket (required)
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket (required)
	KeyFile string `toml:"key_file,omitempty"`

	//RPKIValidator is the path to the shell scripts that takes 2 arguments,
	//ASID and the prefix to validate (required)
	RPKIValidator string `toml:"rpki_validator,omitempty"`

	//RPKIValidString is the response of the validator script if the
	//ASID and prefix are valid (required)
	RPKIValidString string `toml:"rpki_entry_valid,omitempty"`

	//PLNIA IA of the PLN to contact for PCN lists (required)
	PLNIA addr.IA `toml:"pln_isd_as,omitempty"`

	//MSListValidTime time for which a published MS list is
	//valid in minutes (default = 10080) 1 week
	MSListValidTime time.Duration `toml:"ms_list_valid_time,omitempty"`

	//MSPullListInterval time intervaal to pull full MS list in minutes (default = 1440) 1 day
	MSPullListInterval time.Duration ` toml:"ms_pull_list_interval,omitempty"`
}

func (cfg *MsConf) InitDefaults() {
	if cfg.Db == "" {
		cfg.Db = DefaultDb
	}

	if cfg.MSListValidTime == 0 {
		cfg.MSListValidTime = DefaultMSListValidTime
	}

	if cfg.MSPullListInterval == 0 {
		cfg.MSPullListInterval = DefaultMSPullListInterval
	}
}
func (cfg *MsConf) Validate() error {

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
	if cfg.RPKIValidator == "" {
		return serrors.New("rpki_validator should be set")
	}
	if cfg.PLNIA.IsZero() {
		return serrors.New("pln_isd_as must be set")
	}
	if cfg.RPKIValidString == "" {
		return serrors.New("rpki_entry_valid should be set")
	}

	return nil
}

func (cfg *MsConf) ConfigName() string {
	return "ms"
}

func (cfg *MsConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(msSample, ctx[config.ID]))
}
