package backend

import (
	"ufleet-deploy/pkg/log"
)

func Init() error {
	be := NewBackendHandler()
	//创建根key
	err := initRootKey()
	if err != nil {
		return err
	}

	//创建各资源key
	err = initResourcesKey()
	if err != nil {
		return err
	}

	log.DebugPrint("tune resources group according to external group")
	for _, v := range resources {
		err := TuneResourceGroupAccordingToExternalGroup(be, v, nil)
		if err != nil {
			return err
		}

	}

	//必须要在ResourceGroup之后进行.
	log.DebugPrint("tune resources workspace according to external workspace")
	for _, v := range resources {
		err := TuneResourcesWorkspaceAccordingToExternalWorkspace(be, v)
		if err != nil {
			return err
		}
	}

	err = watchExternalGroupChange()
	if err != nil {
		return err
	}

	err = watchWorkspaceChange()
	if err != nil {
		return err
	}

	err = watchBackendEvent()
	if err != nil {
		return err
	}
	return nil

}
