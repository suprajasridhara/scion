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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cs/ifstate"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/prom"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/pgn/internal/pgncmn"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
	pgnconfig "github.com/scionproto/scion/go/pkg/pgn/config"
	"github.com/scionproto/scion/go/pkg/service"
	"github.com/scionproto/scion/go/proto"
)

var (
	cfg pgnconfig.Config
)

const (
	shutdownWaitTimeout = 5 * time.Second
)

func init() {
	flag.Usage = env.Usage
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	fatal.Init()
	env.AddFlags()
	flag.Parse()
	if v, ok := env.CheckFlags(&cfg); !ok {
		return v
	}
	if err := setupBasic(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer log.Flush()
	defer env.LogAppStopped("PGN", cfg.Pgn.ID)
	defer log.HandlePanic()
	if err := validateConfig(); err != nil {
		log.Error("Configuration validation failed", "err", err)
		return 1
	}

	if cfg.Metrics.Prometheus != "" {
		cfg.Metrics.StartPrometheus()
	}

	_, err := setupTopo()
	if err != nil {
		log.Error("PGN setupTopo failed", "err", err)
		return 1
	}

	if err := pgncmn.Init(cfg.Pgn, cfg.SD, cfg.Features); err != nil {
		log.Error("PGN common initialization failed", "err", err)
		return 1
	}

	pgnmsgr.ID = cfg.General.ID
	go func() {
		defer log.HandlePanic()
		pgnmsgr.Msgr.ListenAndServe()
	}()
	defer pgnmsgr.Msgr.CloseServer()

	if err := setupDb(); err != nil {
		log.Error("PGN db initialization failed", "err", err)
		return 1
	}
	defer sqlite.Db.Close()

	// Start HTTP endpoints.
	statusPages := service.StatusPages{
		"info":   service.NewInfoHandler(),
		"config": service.NewConfigHandler(cfg),
	}
	if err := statusPages.Register(http.DefaultServeMux, cfg.Pgn.ID); err != nil {
		log.Error("registering status pages", "err", err)
		return 1
	}

	select {
	case <-fatal.ShutdownChan():
		return 0
	case <-fatal.FatalChan():
		return 1
	}

}

func setupDb() error {
	err := sqlite.New(cfg.Pgn.Db, 1)
	if err != nil {
		return serrors.WrapStr("setting up database", err)
	}
	return nil
}

// setupBasic loads the config from file and initializes logging.
func setupBasic() error {
	// Load and initialize config.
	md, err := toml.DecodeFile(env.ConfigFile(), &cfg)
	if err != nil {
		return serrors.WrapStr("Failed to load config", err, "file", env.ConfigFile())
	}
	if len(md.Undecoded()) > 0 {
		return serrors.New("Failed to load config: undecoded keys", "undecoded", md.Undecoded())
	}
	cfg.InitDefaults()
	if err := log.Setup(cfg.Logging); err != nil {
		return serrors.WrapStr("Failed to initialize logging", err)
	}
	prom.ExportElementID(cfg.Pgn.ID)
	return env.LogAppStarted("PGN", cfg.Pgn.ID)
}

func validateConfig() error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	return nil
}

func setupTopo() (*ifstate.Interfaces, error) {
	topo, err := topology.FromJSONFile(cfg.General.Topology())
	if err != nil {
		return nil, serrors.WrapStr("loading topology", err)
	}

	intfs := ifstate.NewInterfaces(topo.IFInfoMap(), ifstate.Config{})
	itopo.Init(&itopo.Config{
		ID:  cfg.General.ID,
		Svc: proto.ServiceType_pgn,
		Callbacks: itopo.Callbacks{
			OnUpdate: func() {
				intfs.Update(itopo.Get().IFInfoMap())
			},
		},
	})

	if err := itopo.Update(topo); err != nil {
		return nil, serrors.WrapStr("setting initial static topology", err)
	}
	infraenv.InitInfraEnvironment(cfg.General.Topology())
	return intfs, nil
}
