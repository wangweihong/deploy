package app

import (
	"ufleet-deploy/pkg/backend"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitAppController(be)
	if err != nil {
		panic(err.Error())
	}
	//注册事件,一旦etcd后端出现事件,进行清理
	backend.RegisterEventHandler(backend.ResourceApps, sm)

	go ResourceEventHandler()

}
