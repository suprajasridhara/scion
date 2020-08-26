// Copyright 2019 ETH Zurich
// Copyright 2019 ETH Zurich, Anapaya Systems
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

package egress

import (
	"context"
	"fmt"
	"io"

	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/sigjson"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/sig/egress/asmap"
	"github.com/scionproto/scion/go/sig/egress/iface"
	"github.com/scionproto/scion/go/sig/egress/reader"
	"github.com/scionproto/scion/go/sig/internal/sigcmn"
)

func Init(tunIO io.ReadWriteCloser) {
	fatal.Check()
	iface.Init()
	// Spawn egress reader
	go func() {
		defer log.HandlePanic()
		reader.NewReader(tunIO).Run()
	}()
}

func ReloadConfig(cfg *sigjson.Cfg) bool {
	/***
	TODO_SIG:(supraja) pull from MS here to get the mapping.
	In the new infrastructure the MS replies with (IP,AS) pairs so change the way the mapping is
	handled here
	***/
	//paths, err := sdConn.Paths(context.Background(), remote.IA, local.IA, sd.PathReqFlags{})

	//TODO (supraja): get core AS here
	dst, _ := snet.ParseUDPAddr("2-ff00:0:221,[127.0.0.133]:30755")

	pathSet := sigcmn.PathMgr.Query(context.Background(), sigcmn.Network.LocalIA, dst.IA, sciond.PathReqFlags{})
	path := pathSet.GetAppPath(snet.PathFingerprint(""))
	fmt.Printf("Using path:\n  %s\n", fmt.Sprintf("%s", path))

	//TODO (supraja): depending on communication mechanism cpnproto/grpc add code here

	res := asmap.Map.ReloadConfig(cfg)

	log.Info("Config reloaded")
	return res
}
