package hpa

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceHorizontalPodAutoscalers
	resourceKind = "HorizontalPodAutoscaler"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitHorizontalPodAutoscalerController(be)
	if err != nil {
		panic(err.Error())
	}

	backend.RegisterEventHandler(backendKind, rm)
	err = resource.RegisterResourceController(resourceKind, rm)
	if err != nil {
		panic(err.Error())
	}

	go HandleEventWatchFromK8sCluster(cluster.HPAEventChan, resourceKind, rm)
}
