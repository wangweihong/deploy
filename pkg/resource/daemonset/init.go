package daemonset

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceDaemonSets
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitDaemonSetController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
