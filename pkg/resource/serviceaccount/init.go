package serviceaccount

import "ufleet-deploy/pkg/backend"

const (
	backendKind = backend.ResourceServiceAccounts
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitServiceAccountController(be)
	if err != nil {
		panic(err.Error())
	}
	backend.RegisterEventHandler(backendKind, EventHandler)

	go HandleClusterResourceEvent()
}
