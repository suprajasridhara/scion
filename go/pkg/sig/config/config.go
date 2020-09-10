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
	DefaultCtrlPort    = 30256
	DefaultEncapPort   = 30056
	DefaultTunName     = "sig"
	DefaultTunRTableId = 11
)

type Config struct {
	General  env.General `toml:"general ,omitempty"`
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	Sciond   env.SCIONDClient `toml:"sciond_connection,omitempty"`
	Sig      SigConf          `toml:"sig,omitempty"`
}

func (cfg *Config) InitDefaults() {
	config.InitAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Sig,
	)
}

func (cfg *Config) Validate() error {
	return config.ValidateAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Sig,
	)
}

func (cfg *Config) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	config.WriteSample(dst, path, config.CtxMap{config.ID: idSample},
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Sig,
	)
}

var _ config.Config = (*SigConf)(nil)

// SigConf contains the configuration specific to the SIG.
type SigConf struct {
	// ID of the SIG (required)
	ID string `toml:"id,omitempty"`
	// The SIG config json file. (required)
	SIGConfig string `toml:"sig_config,omitempty"`
	// IA the local IA (required)
	IA addr.IA `toml:"isd_as,omitempty"`
	// IP the bind IP address (required)
	IP net.IP `toml:"ip,omitempty"`
	// Control data port, e.g. keepalives. (default DefaultCtrlPort)
	CtrlPort uint16 `toml:"ctrl_port,omitempty"`
	// Encapsulation data port. (default DefaultEncapPort)
	EncapPort uint16 `toml:"encap_port,omitempty"`
	// Name of TUN device to create. (default DefaultTunName)
	Tun string `toml:"tun,omitempty"`
	// TunRTableId the id of the routing table used in the SIG. (default DefaultTunRTableId)
	TunRTableId int `toml:"tun_routing_table_id,omitempty"`
	// IPv4 source address hint to put into routing table.
	SrcIP4 net.IP `toml:"src_ipv4,omitempty"`
	// IPv6 source address hint to put into routing table.
	SrcIP6 net.IP `toml:"src_ipv6,omitempty"`
	// DispatcherBypass is the underlay address (e.g. ":30041") to use when bypassing SCION
	// dispatcher. If the field is empty bypass is not done and SCION dispatcher is used
	// instead.
	DispatcherBypass string `toml:"disaptcher_bypass,omitempty"`

	//config directory to read crypto keys from
	CfgDir string `toml:"cfg_dir,omitempty"`

	//db to store sig cfg data (default ./sig.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//UDP port to open a messenger connection on
	UDPPort uint16 `toml:"udp_port,omitempty"`

	//QUIC IP:Port
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket
	KeyFile string `toml:"key_file,omitempty"`

	//PrefixFile contains the list of prefixes that should be pushed to a Mapping service in the ISD. This file is scanned periodically for changes
	PrefixFile string `toml:"prefix_file,omitempty"`

	//PrefixPushInterval in minutes is the interval between 2 consecutive pushes of prefixes to the mapping service. default (60)
	PrefixPushInterval time.Duration `toml:"prefix_push_interval,omitempty"`
}

// InitDefaults sets the default values to unset values.
func (cfg *SigConf) InitDefaults() {
}

// Validate validate the config and returns an error if a value is not valid.
func (cfg *SigConf) Validate() error {
	if cfg.ID == "" {
		return serrors.New("id must be set!")
	}
	if cfg.SIGConfig == "" {
		return serrors.New("sig_config must be set!")
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
	if cfg.CfgDir == "" {
		return serrors.New("sig cfg_dir should be set")
	}
	if cfg.PrefixFile == "" {
		return serrors.New("prefix_file should be set")
	}
	if cfg.CtrlPort == 0 {
		cfg.CtrlPort = DefaultCtrlPort
	}
	if cfg.EncapPort == 0 {
		cfg.EncapPort = DefaultEncapPort
	}
	if cfg.Tun == "" {
		cfg.Tun = DefaultTunName
	}
	if cfg.TunRTableId == 0 {
		cfg.TunRTableId = DefaultTunRTableId
	}
	if cfg.Db == "" {
		cfg.Db = "/sig.db"
	}
	if cfg.PrefixPushInterval == 0 {
		cfg.PrefixPushInterval = 60
	}
	return nil
}

func (cfg *SigConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(sigSample, ctx[config.ID]))
}

func (cfg *SigConf) ConfigName() string {
	return "sig"
}
