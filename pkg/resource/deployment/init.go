package deployment

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceDeployments
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitDeploymentController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
