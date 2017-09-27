package init

import (
	"fmt"
	"os"
	"ufleet-deploy/pkg/app"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
)

const (
	etcdHostEnvKey = "ETCDHOST"
)

func init() {
	log.DebugPrint("init kv controller")
	initKV()
	log.DebugPrint("init backend controller")
	backend.Init()
	log.DebugPrint("init app controller")
	app.Init()

}

func initKV() {
	etcdHost := os.Getenv(etcdHostEnvKey)
	if len(etcdHost) == 0 {
		panic(fmt.Sprintf("must provide Environment \"%v\"", etcdHostEnvKey))
	}
	kv.Init(etcdHost)
}
