package service

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceServices
	resourceKind = "Service"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitServiceController(be)
	if err != nil {
		panic(err.Error())
	}
	err = resource.RegisterCURInterface(resourceKind, Controller)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
