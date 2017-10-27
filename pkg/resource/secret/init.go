package secret

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceSecrets
	resourceKind = "Secret"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitSecretController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)

	go resource.HandleEventWatchFromK8sCluster(cluster.SecretEventChan, resourceKind, rm)
}
