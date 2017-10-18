package app

import (
	"ufleet-deploy/pkg/backend"
)

func Init() {
	be := backend.NewBackendHandler()
	//	storer = InitStore()

	/*
		err := backend.RegisterWorkspaceNoticer(backend.ResourceApps)
		if err != nil {
			panic(err.Error())
		}

		err := backend.RegisterExternalGroupNoticer(backend.ResourceApps)
		if err != nil {
			panic(err.Error())
		}
	*/

	var err error
	Controller, err = InitAppController(be)
	if err != nil {
		panic(err.Error())
	}
	//注册事件,一旦etcd后端出现事件,进行清理
	backend.RegisterEventHandler(backend.ResourceApps, EventHandler)

	go ResourceEventHandler()

}
