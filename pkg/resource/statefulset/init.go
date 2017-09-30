package statefulset

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceStatefulSets
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitStatefulSetController(be)
	if err != nil {
		panic(err.Error())
	}

	go HandleClusterResourceEvent()
}
