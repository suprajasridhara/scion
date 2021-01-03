package config_test

import (
	"bytes"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"

	"github.com/scionproto/scion/go/lib/env/envtest"
	"github.com/scionproto/scion/go/lib/log/logtest"
	"github.com/scionproto/scion/go/pkg/pln/config"
	"github.com/scionproto/scion/go/pkg/pln/config/configtest"
)

func TestConfigSample(t *testing.T) {
	var sample bytes.Buffer
	var cfg config.Config
	cfg.Sample(&sample, nil, nil)

	InitTestConfig(&cfg)
	err := toml.NewDecoder(bytes.NewReader(sample.Bytes())).Strict(true).Decode(&cfg)
	assert.NoError(t, err)
	CheckTestConfig(t, &cfg, "pln1")
}

func InitTestConfig(cfg *config.Config) {
	envtest.InitTest(nil, &cfg.Metrics, nil, &cfg.Sciond)
	logtest.InitTestLogging(&cfg.Logging)
}

func CheckTestConfig(t *testing.T, cfg *config.Config, id string) {
	envtest.CheckTest(t, nil, &cfg.Metrics, nil, &cfg.Sciond, id)
	logtest.CheckTestLogging(t, &cfg.Logging, id)
	configtest.CheckTestPLN(t, &cfg.Pln, id)
}
