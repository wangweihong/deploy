package serviceaccount

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceServiceAccounts
	resourceKind = "ServiceAccount"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitServiceAccountController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)
	err = resource.RegisterResourceController(resourceKind, rm)
	if err != nil {
		panic(err.Error())
	}

	go resource.HandleEventWatchFromK8sCluster(cluster.ServiceAccountEventChan, resourceKind, rm)
}
