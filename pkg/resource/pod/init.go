package pod

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourcePods
	resourceKind = "Pod"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitPodController(be)
	if err != nil {
		panic(err.Error())
	}

	backend.RegisterEventHandler(backendKind, rm)
	err = resource.RegisterResourceController(resourceKind, rm)
	if err != nil {
		panic(err.Error())
	}

	go resource.HandleEventWatchFromK8sCluster(cluster.PodEventChan, resourceKind, rm)
}
