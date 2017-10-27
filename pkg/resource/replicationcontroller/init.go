package replicationcontroller

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceReplicationControllers
	resourceKind = "ReplicationController"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitReplicationControllerController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)

	go resource.HandleEventWatchFromK8sCluster(cluster.ReplicationControllerEventChan, resourceKind, rm)
}
