package configmap

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceConfigMaps
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitConfigMapController(be)
	if err != nil {
		panic(err.Error())
	}

	go HandleClusterResourceEvent()
}
