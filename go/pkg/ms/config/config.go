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
	//TODO_MS:(supraja) create the config definition to start the MS instance
	//sciond config, IP, PLN config
	ID               string  `toml:"id,omitempty"`
	DispatcherBypass string  `toml:"disaptcher_bypass,omitempty"`
	IP               net.IP  `toml:"ip,omitempty"`
	CtrlPort         uint16  `toml:"ctrl_port,omitempty"`
	DataPort         uint16  `toml:"data_port,omitempty"`
	IA               addr.IA `toml:"isd_as,omitempty"`
	Address          string  `toml:"address,omitempty"`

	//config directory to read crypto keys from
	CfgDir string `toml:"cfg_dir,omitempty"`

	//db to store ms cfg data (default ./ms.db will be created or read from)
	Db string `toml:"db,omitempty"`

	//UDP port to open a messenger connection on
	UDPPort uint16 `toml:"udp_port,omitempty"`

	//QUIC IP:Port
	QUICAddr string `toml:"quic_addr,omitempty"`

	//CertFile for QUIC socket
	CertFile string `toml:"cert_file,omitempty"`

	//KeyFile for QUIC socket
	KeyFile string `toml:"key_file,omitempty"`

	//RPKIValidator is the path to the shell scripts that takes 2 arguments,
	//ASID and the prefix to validate
	RPKIValidator string `toml:"rpki_validator,omitempty"`

	//RPKIValidString is the response of the validator script if the ASID and prefix are valid
	RPKIValidString string `toml:"rpki_entry_valid,omitempty"`

	PLNIA addr.IA `toml:"pln_isd_as,omitempty"`
}

func (cfg *MsConf) InitDefaults() {
	//TODO_MS:(supraja)

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
	if cfg.CfgDir == "" {
		return serrors.New("ms cfg_dir should be set")
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
	if cfg.Db == "" {
		cfg.Db = "/ms.db"
	}
	return nil
}

func (cfg *MsConf) ConfigName() string {
	return "ms"
}

func (cfg *MsConf) Sample(dst io.Writer, path config.Path, ctx config.CtxMap) {
	config.WriteString(dst, fmt.Sprintf(msSample, ctx[config.ID]))
}
