package init

import (
	"fmt"
	"os"
	"ufleet-deploy/pkg/app"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
)

const (
	etcdHostEnvKey    = "ETCDHOST"
	clusterHostEnvKey = "CLUSTERHOST"
)

func init() {
	log.DebugPrint("init kv controller")
	initKV()
	log.DebugPrint("init backend controller")
	backend.Init()
	log.DebugPrint("init app controller")
	app.Init()

	initCluster()

}

func initKV() {
	etcdHost := os.Getenv(etcdHostEnvKey)
	if len(etcdHost) == 0 {
		panic(fmt.Sprintf("must provide Environment \"%v\"", etcdHostEnvKey))
	}
	kv.Init(etcdHost)
}

func initCluster() {
	clusterHost := os.Getenv(clusterHostEnvKey)
	if len(clusterHost) == 0 {
		panic(fmt.Sprintf("must provide Environment \"%v\"", clusterHostEnvKey))
	}

	cluster.Init(clusterHost)
}
