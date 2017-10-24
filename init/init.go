package init

import (
	"fmt"
	"os"
	"ufleet-deploy/pkg/app"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/configmap"
	"ufleet-deploy/pkg/resource/cronjob"
	"ufleet-deploy/pkg/resource/daemonset"
	"ufleet-deploy/pkg/resource/deployment"
	"ufleet-deploy/pkg/resource/endpoint"
	"ufleet-deploy/pkg/resource/ingress"
	"ufleet-deploy/pkg/resource/job"
	"ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/replicaset"
	"ufleet-deploy/pkg/resource/replicationcontroller"
	"ufleet-deploy/pkg/resource/secret"
	"ufleet-deploy/pkg/resource/service"
	"ufleet-deploy/pkg/resource/serviceaccount"
	"ufleet-deploy/pkg/resource/statefulset"
)

const (
	etcdHostEnvKey        = "ETCDHOST"
	clusterHostEnvKey     = "CLUSTERHOST"
	clusterCurrentHostKey = "CURRENT_HOST"
)

func init() {
	log.DebugPrint("init kv controller")
	initKV()
	log.DebugPrint("init backend controller")
	backend.Init()
	log.DebugPrint("init app controller")
	app.Init()
	log.DebugPrint("init pod controller")
	pod.Init()

	service.Init()

	secret.Init()
	configmap.Init()
	serviceaccount.Init()
	endpoint.Init()
	deployment.Init()
	daemonset.Init()
	ingress.Init()
	statefulset.Init()
	job.Init()
	cronjob.Init()
	replicationcontroller.Init()
	replicaset.Init()

	//需要在pod/service等resource后初始化
	//因为初始化就构建k8s的对象到内存中
	log.DebugPrint("init cluster controller")
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
	currentHost := os.Getenv(clusterCurrentHostKey)
	if len(clusterHost) == 0 {
		panic(fmt.Sprintf("must provide Environment \"%v\"", clusterHostEnvKey))
	}

	cluster.Init(clusterHost, currentHost)
}
