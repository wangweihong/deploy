package job

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceJobs
	resourceKind = "Job"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitJobController(be)
	if err != nil {
		panic(err.Error())
	}
	err = resource.RegisterCURInterface(resourceKind, Controller)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, rm)

	go HandleClusterResourceEvent()
}
