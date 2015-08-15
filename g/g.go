package g

import (
	"runtime"
)

// changelog:
// 1.3.2: add config file `graph_bachends.txt`
// 1.3.2.sdp_v001 add graph/sdp/one api for console
// TODO: mv graph cluster config to cfg.json

const (
	VERSION = "1.3.2.sdp_v001"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
