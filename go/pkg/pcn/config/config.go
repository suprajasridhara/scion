package config

import (
	"fmt"
	"io"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/config"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
)

type Config struct {
	General  env.General `toml:"general,omitempty"`
	Features env.Features
	Logging  log.Config       `toml:"log,omitempty"`
	Metrics  env.Metrics      `toml:"metrics,omitempty"`
	Sciond   env.SCIONDClient `toml:"sciond_connection,omitempty"`
	Pcn      PcnConf          `toml:"pcn,omitempty"`
}

func (cfg *Config) InitDefaults() {
	config.InitAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pcn,
	)
}

func (cfg *Config) Validate() error {
	return config.ValidateAll(
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pcn,
	)
}

func (cfg *Config) Sample(dst io.Writer, path config.Path, _ config.CtxMap) {
	print(path)
	config.WriteSample(dst, path, config.CtxMap{config.ID: idSample},
		&cfg.Features,
		&cfg.Logging,
		&cfg.Metrics,
		&cfg.Sciond,
		&cfg.Pcn,
	)
}

var _ config.Config = (*PcnConf)(nil)

type PcnConf struct {
	ID               string  `toml:"id,omitempty"`
	DispatcherBypass string  `toml:"disaptcher_bypass,omitempty"`
	IP               net.IP  `toml:"ip,omitempty"`
	CtrlPort         uint16  `toml:"ctrl_port,omitempty"`
	IA               addr.IA `toml:"isd_as,omitempty"`
	Address          string  `toml:"address,omitempty"`

	//config directory to read crypto keys from
	CfgDir string `toml:"cfg_dir,omitempty"`

	//db to store pcn cfg data (default ./pcn.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//UDP port to open a messenger connection on
	UDPPort uint16 `toml:"udp_port,omitempty"`

	//QUIC IP:Port
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket
	KeyFile string `toml:"key_file,omitempty"`

	PLNIA addr.IA `toml:"pln_isd_as,omitempty"`
}

func (cfg *PcnConf) InitDefaults() {
	//TODO (supraja): set this if needed

}
func (cfg *PcnConf) Validate() error {

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
	if cfg.CfgDir == "" {
		return serrors.New("pcn cfg_dir should be set")
	}
	if cfg.Db == "" {
		cfg.Db = "/pcn.db"
	}
	if cfg.PLNIA.IsZero() {
		return serrors.New("pln_isd_as must be set")
	}
	return nil
}

func (cfg *PcnConf) ConfigName() string {
	return "pcn"
}

func (cfg *PcnConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(pcnSample, ctx[config.ID]))
}
