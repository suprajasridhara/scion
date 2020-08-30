// Copyright 2017 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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
	"bytes"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cs/ifstate"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/prom"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/pkg/service"
	"github.com/scionproto/scion/go/proto"

	"github.com/scionproto/scion/go/ms/internal/mscmn"
	msconfig "github.com/scionproto/scion/go/pkg/ms/config"
)

var (
	cfg msconfig.Config
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
	defer env.LogAppStopped("MS", cfg.Ms.ID)
	defer log.HandlePanic()
	if err := validateConfig(); err != nil {
		log.Error("Configuration validation failed", "err", err)
		return 1
	}

	//TODO_MS:(supraja) init MS here
	env.SetupEnv(
		func() {
			success := loadConfig("nil") //TODO_MS:(supraja) remove if we dont need this
			// Errors already logged in loadConfig
			log.Info("reloadOnSIGHUP: reload done", "success", success)
		},
	)

	//TODO_MS:(supraja) remove if we dont need this
	if loadConfig("") != true {
		log.Error("MS configuration loading failed")
		return 1
	}
	cfg.Metrics.StartPrometheus()
	intfs, err := setupTopo()
	if err != nil {
		log.Error("MS setupTopo failed", "err", err)
		return 1
	}
	if err := mscmn.Init(cfg.Ms, cfg.Sciond, cfg.Features); err != nil {
		log.Error("MS common initialization failed", "err", err)
		return 1
	}
	// Keepalive mechanism is deprecated and will be removed with change to
	// header v2. Disable with https://github.com/Anapaya/scion/issues/3337.
	if !cfg.Features.HeaderV2 || true {
		mscmn.Msgr.AddHandler(infra.IfStateReq, ifstate.NewHandler(intfs))
		//TODO (supraja): fix this with a handler that works if needed
		mscmn.Msgr.AddHandler(infra.IfId, mscmn.IdIdHandler{})
	}

	go func() {
		defer log.HandlePanic()
		mscmn.Msgr.ListenAndServe()
	}()
	defer mscmn.Msgr.CloseServer()
	// Start HTTP endpoints.
	statusPages := service.StatusPages{
		"info":   service.NewInfoHandler(),
		"config": service.NewConfigHandler(cfg),
	}
	if err := statusPages.Register(http.DefaultServeMux, cfg.Ms.ID); err != nil {
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

func setupTopo() (*ifstate.Interfaces, error) {
	//itopo.Init(&itopo.Config{})

	topo, err := topology.FromJSONFile(cfg.General.Topology())
	if err != nil {
		return nil, serrors.WrapStr("loading topology", err)
	}

	intfs := ifstate.NewInterfaces(topo.IFInfoMap(), ifstate.Config{})
	//prometheus.MustRegister(ifstate.NewCollector(intfs))
	itopo.Init(&itopo.Config{
		ID:  cfg.General.ID,
		Svc: proto.ServiceType_ms,
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
	prom.ExportElementID(cfg.Ms.ID)
	return env.LogAppStarted("MS", cfg.Ms.ID)
}

func validateConfig() error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Metrics.Prometheus == "" {
		cfg.Metrics.Prometheus = "127.0.0.1:1281"
	}
	return nil
}

func loadConfig(path string) bool {
	//TODO_MS:(supraja) do we need a config file?
	return true
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	print("configHandler")
	w.Header().Set("Content-Type", "text/plain")
	var buf bytes.Buffer
	toml.NewEncoder(&buf).Encode(cfg)
	fmt.Fprint(w, buf.String())
}