package cronjob

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceCronJobs
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitCronJobController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
