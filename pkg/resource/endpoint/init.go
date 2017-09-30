package endpoint

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceEndpoints
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitEndpointController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
