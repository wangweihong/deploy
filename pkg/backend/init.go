package backend

import (
	"ufleet-deploy/pkg/log"
)

func Init() {
	be := NewBackendHandler()
	err := initRootKey()
	if err != nil {
		panic(err.Error())
	}

	log.DebugPrint("tune resources group according to external group")

	for _, v := range resources {
		err := TuneResourceGroupAccordingToExternalGroup(be, v, nil)
		if err != nil {
			panic(err.Error())
		}

	}

	//必须要在ResourceGroup之后进行.
	log.DebugPrint("tune resources workspace according to external workspace")
	for _, v := range resources {
		err := TuneResourcesWorkspaceAccordingToExternalWorkspace(be, v)
		if err != nil {
			panic(err.Error())
		}

	}
	/*
		err = watchRegistyEvent()
		if err != nil {
			panic(err.Error())
		}
	*/

	err = watchExternalGroupChange()
	if err != nil {
		panic(err.Error())
	}

	err = watchWorkspaceChange()
	if err != nil {
		panic(err.Error())
	}

	err = watchBackendEvent()
	if err != nil {
		panic(err.Error())
	}

}
