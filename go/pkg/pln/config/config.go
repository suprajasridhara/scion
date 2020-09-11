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
	ID               string  `toml:"id,omitempty"`
	DispatcherBypass string  `toml:"disaptcher_bypass,omitempty"`
	IP               net.IP  `toml:"ip,omitempty"`
	CtrlPort         uint16  `toml:"ctrl_port,omitempty"`
	IA               addr.IA `toml:"isd_as,omitempty"`
	Address          string  `toml:"address,omitempty"`

	//config directory to read crypto keys from
	CfgDir string `toml:"cfg_dir,omitempty"`

	//db to store pln cfg data (default ./pln.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//UDP port to open a messenger connection on
	UDPPort uint16 `toml:"udp_port,omitempty"`

	//QUIC IP:Port
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket
	KeyFile string `toml:"key_file,omitempty"`
}

func (cfg *PlnConf) InitDefaults() {
	//TODO (supraja): set this if needed

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
	if cfg.CfgDir == "" {
		return serrors.New("pln cfg_dir should be set")
	}
	if cfg.Db == "" {
		cfg.Db = "/pln.db"
	}
	return nil
}

func (cfg *PlnConf) ConfigName() string {
	return "pln"
}

func (cfg *PlnConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(plnSample, ctx[config.ID]))
}
