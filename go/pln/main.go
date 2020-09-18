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
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/prom"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/topology"
	plnconfig "github.com/scionproto/scion/go/pkg/pln/config"
	"github.com/scionproto/scion/go/pkg/service"
	"github.com/scionproto/scion/go/pln/internal/plncmn"
	"github.com/scionproto/scion/go/pln/internal/plnmsgr"
	"github.com/scionproto/scion/go/pln/internal/sqlite"
	"github.com/scionproto/scion/go/proto"
)

var (
	cfg plnconfig.Config
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
	defer env.LogAppStopped("PLN", cfg.Pln.ID)
	defer log.HandlePanic()
	if err := validateConfig(); err != nil {
		log.Error("Configuration validation failed", "err", err)
		return 1
	}

	cfg.Metrics.StartPrometheus()
	intfs, err := setupTopo()
	if err != nil {
		log.Error("PLN setupTopo failed", "err", err)
		return 1
	}
	if err := plncmn.Init(cfg.Pln, cfg.Sciond, cfg.Features); err != nil {
		log.Error("PLN common initialization failed", "err", err)
		return 1
	}
	// Keepalive mechanism is deprecated and will be removed with change to
	// header v2. Disable with https://github.com/Anapaya/scion/issues/3337.
	if !cfg.Features.HeaderV2 || true {
		plnmsgr.Msgr.AddHandler(infra.IfStateReq, ifstate.NewHandler(intfs))
		plnmsgr.Msgr.AddHandler(infra.IfId, plncmn.IfIdHandler{})
	}

	go func() {
		defer log.HandlePanic()
		plnmsgr.Msgr.ListenAndServe()
	}()

	setupDb()

	defer plnmsgr.Msgr.CloseServer()
	// Start HTTP endpoints.
	statusPages := service.StatusPages{
		"info":   service.NewInfoHandler(),
		"config": service.NewConfigHandler(cfg),
	}
	if err := statusPages.Register(http.DefaultServeMux, cfg.Pln.ID); err != nil {
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
	err := sqlite.New(cfg.Pln.Db, 1)
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
	prom.ExportElementID(cfg.Pln.ID)
	return env.LogAppStarted("PLN", cfg.Pln.ID)
}

func validateConfig() error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Metrics.Prometheus == "" {
		cfg.Metrics.Prometheus = "127.0.0.1:1282"
	}
	return nil
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
		Svc: proto.ServiceType_pln,
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
