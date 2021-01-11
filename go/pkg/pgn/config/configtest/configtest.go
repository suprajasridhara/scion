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

package configtest

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scionproto/scion/go/lib/xtest"
	"github.com/scionproto/scion/go/pkg/pgn/config"
)

func CheckTestPGN(t *testing.T, cfg *config.PGNConf, id string) {
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
