package configtest

import (
	"net"
	"testing"

	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/pkg/ms/config"
	"github.com/stretchr/testify/assert"
)

func CheckTestPGN(t *testing.T, cfg *config.PgnConf, id string) {
	assert.Equal(t, id, cfg.ID)
	assert.Equal(t, xtest.MustParseIA("1-ff00:0:110"), cfg.IA)
	assert.Equal(t, net.ParseIP("127.0.0.65"), cfg.IP)
	assert.Equal(t, 3009, int(cfg.Port))
	assert.Equal(t, "gen/ISD1/ASff00_0_110", cfg.CfgDir)
	assert.Equal(t, config.DefaultDb, cfg.Db)
	assert.Equal(t, "127.0.0.27:20655", cfg.QUICAddr)
	assert.Equal(t, "gen-certs/tls.pem", cfg.CertFile)
	assert.Equal(t, "gen-certs/tls.key", cfg.KeyFile)
	assert.Equal(t, xtest.MustParseIA("1-ff00:0:110"), cfg.PLNIA)
}
