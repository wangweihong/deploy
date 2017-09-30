package ingress

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceIngresss
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitIngressController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
