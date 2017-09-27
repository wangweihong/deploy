package app

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
)

type AppController interface {
	NewApp(group, workspace, app, describe string, opt CreateOptions) error
	DeleteApp(group, workspace, app string, opt DeleteOptions) error
	AppGetter
	AppLister
}

type Backend interface {
	Create(Data string) error
	Update(Data string) error
	Remove() error
}

type AppMananger struct {
	Groups map[string]AppGroup `json:"groups"`
	Locker Locker
	BE     Backend
}

func InitAppController(be backend.BackendHandler) (AppController, error) {
	sm = &AppMananger{}
	sm.Groups = make(map[string]AppGroup)
	sm.Locker = &sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group AppGroup
		group.Workspaces = make(map[string]AppWorkspace)
		for i, j := range v.Workspaces {
			var workspace AppWorkspace
			workspace.Apps = make(map[string]App)
			for m, n := range j.Resources {
				var app App
				err := json.Unmarshal([]byte(n), &app)
				if err != nil {
					return nil, fmt.Errorf("init app manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Apps[m] = app
			}
			group.Workspaces[i] = workspace
		}
		sm.Groups[k] = group
	}
	return sm, nil
}

type ListOptions struct {
	Workspace *string
}

type CreateOptions struct {
	MemoryOnly string
}

type DeleteOptions struct {
	WaitToComplete bool
	MemoryOnly     bool //不处理存储后端数据
}

type Locker interface {
	Lock()
	Unlock()
}

func (sm *AppMananger) NewApp(groupName, workspaceName, appName string, desc string, opt CreateOptions) error {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()
	_, err := sm.AppIF(groupName, workspaceName, appName)
	if err == nil || IsAppNotFound(err) {
		return ErrAppExists
	}
	//加锁
	//
	return nil

}

func (sm *AppMananger) AppIF(groupName, workspaceName, name string) (AppInterface, error) {
	var opt ListOptions
	opt.Workspace = &workspaceName

	sis, err := sm.AppIFs(groupName, opt)
	if err != nil {
		return nil, err
	}

	for _, v := range sis {
		app := v.Info()
		if app.Name == name {
			return v, nil
		}
	}

	return nil, ErrAppNotFound
}

func (sm *AppMananger) AppIFs(groupName string, opt ListOptions) ([]AppInterface, error) {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	sis := make([]AppInterface, 0)

	group, ok := sm.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	if opt.Workspace != nil {
		workspace, ok := group.Workspaces[*opt.Workspace]
		if !ok {
			return nil, ErrWorkspaceNotFound
		}

		for _, v := range workspace.Apps {
			sis = append(sis, &v)
		}
		return sis, nil
	}

	for _, v := range group.Workspaces {
		for _, j := range v.Apps {
			sis = append(sis, &j)
		}
	}
	return sis, nil

}

func (sm *AppMananger) DeleteApp(groupName, workspaceName, name string, opt DeleteOptions) error {
	si, err := sm.AppIF(groupName, workspaceName, name)
	if err != nil {
		return err
	}

	app := si.Info()

	for _, v := range app.Resources {

		key := generateResourceKey(v.Kind, v.Name)
		err := si.RemoveResource(v.Kind, v.Name, false)
		if err != nil {
			err2 := storer.Update(groupName, workspaceName, v.Name, app)
			if err2 != nil {
				log.DebugPrint("store to app backend fail for %v", err)
			}
			return err
		}
		delete(app.Resources, key)
	}
	//删应用
	if !opt.MemoryOnly {
		err := storer.Delete(groupName, workspaceName, app.Name)
		if err != nil {
			return log.DebugPrint(err)
		}
	}

	return nil
}
