package job

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceJobs
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitJobController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
