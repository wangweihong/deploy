package pod

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourcePods
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitPodController(be)
	if err != nil {
		panic(err.Error())
	}

	go HandleClusterResourceEvent()
}
