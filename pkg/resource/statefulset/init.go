package statefulset

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceStatefulSets
	resourceKind = "StatefulSet"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitStatefulSetController(be)
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
