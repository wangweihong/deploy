package replicationcontroller

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceReplicationControllers
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitReplicationControllerController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
