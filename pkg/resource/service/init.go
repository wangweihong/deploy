package service

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceServices
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitServiceController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
