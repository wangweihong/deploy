package ingress

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceIngresss
	resourceKind = "Ingress"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitIngressController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)
	err = resource.RegisterResourceController(resourceKind, rm)
	if err != nil {
		panic(err.Error())
	}

	go resource.HandleEventWatchFromK8sCluster(cluster.IngressEventChan, resourceKind, rm)
}
