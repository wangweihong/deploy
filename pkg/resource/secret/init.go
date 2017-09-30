package secret

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceSecrets
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitSecretController(be)
	if err != nil {
		panic(err.Error())
	}

	go HandleClusterResourceEvent()
}
