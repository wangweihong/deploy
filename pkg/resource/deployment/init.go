package deployment

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceDeployments
	resourceKind = "Deployment"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitDeploymentController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)

	go resource.HandleEventWatchFromK8sCluster(cluster.DeploymentEventChan, resourceKind, rm)
}
