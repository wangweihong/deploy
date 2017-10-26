package configmap

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceConfigMaps
	resourceKind = "ConfigMap"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitConfigMapController(be)
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
