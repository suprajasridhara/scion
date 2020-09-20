package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/scionproto/scion/go/cs/ifstate"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/env"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/infra/modules/itopo"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/prom"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/pcn/internal/pcncmn"
	"github.com/scionproto/scion/go/pcn/internal/pcnmsgr"
	"github.com/scionproto/scion/go/pcn/internal/sqlite"
	"github.com/scionproto/scion/go/pcn/plncomm"
	pcnconfig "github.com/scionproto/scion/go/pkg/pcn/config"
	"github.com/scionproto/scion/go/pkg/service"
	"github.com/scionproto/scion/go/proto"
)

var (
	cfg pcnconfig.Config
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
	defer env.LogAppStopped("PCN", cfg.Pcn.ID)
	defer log.HandlePanic()
	if err := validateConfig(); err != nil {
		log.Error("Configuration validation failed", "err", err)
		return 1
	}

	cfg.Metrics.StartPrometheus()
	intfs, err := setupTopo()
	if err != nil {
		log.Error("PCN setupTopo failed", "err", err)
		return 1
	}
	if err := pcncmn.Init(cfg.Pcn, cfg.Sciond, cfg.Features); err != nil {
		log.Error("PCN common initialization failed", "err", err)
		return 1
	}
	// Keepalive mechanism is deprecated and will be removed with change to
	// header v2. Disable with https://github.com/Anapaya/scion/issues/3337.
	if !cfg.Features.HeaderV2 || true {
		pcnmsgr.Msgr.AddHandler(infra.IfStateReq, ifstate.NewHandler(intfs))
		pcnmsgr.Msgr.AddHandler(infra.IfId, pcncmn.IfIdHandler{})
	}

	pcnmsgr.Id = cfg.General.ID
	go func() {
		defer log.HandlePanic()
		pcnmsgr.Msgr.ListenAndServe()
	}()

	if err := setupDb(); err != nil {
		log.Error("PCN db initialization failed", "err", err)
		return 1
	}

	go func(ctx context.Context, pcnId string, ia addr.IA, plnIA addr.IA) {
		defer log.HandlePanic()
		plncomm.AddPCNEntry(ctx, pcnId, ia, plnIA)
	}(context.Background(), cfg.General.ID, pcncmn.IA, pcncmn.PLNIA)

	defer pcnmsgr.Msgr.CloseServer()
	// Start HTTP endpoints.
	statusPages := service.StatusPages{
		"info":   service.NewInfoHandler(),
		"config": service.NewConfigHandler(cfg),
	}
	if err := statusPages.Register(http.DefaultServeMux, cfg.Pcn.ID); err != nil {
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
	err := sqlite.New(cfg.Pcn.Db, 1)
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
	prom.ExportElementID(cfg.Pcn.ID)
	return env.LogAppStarted("PCN", cfg.Pcn.ID)
}

func validateConfig() error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Metrics.Prometheus == "" {
		cfg.Metrics.Prometheus = "127.0.0.1:1289"
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
		Svc: proto.ServiceType_pcn,
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
