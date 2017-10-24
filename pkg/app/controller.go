package app

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
)

type AppController interface {
	NewApp(group, workspace, app string, describe []byte, opt CreateOption) error
	DeleteApp(group, workspace, app string, opt DeleteOption) error
	Get(group, workspaceName, name string) (AppInterface, error)
	List(group string, opt ListOption) ([]AppInterface, error)
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

type ListOption struct {
	Workspace *string
}

type CreateOption struct {
	User string
}

type DeleteOption struct {
	WaitToComplete bool
}

type Locker interface {
	Lock()
	Unlock()
}

func (sm *AppMananger) NewApp(groupName, workspaceName, appName string, desc []byte, opt CreateOption) error {

	sm.Locker.Lock()
	_, err := sm.get(groupName, workspaceName, appName)
	if err == nil {
		sm.Locker.Unlock()
		return ErrResourceExists
	}
	//加锁
	//
	var stack App
	stack.Name = appName
	stack.Group = groupName
	stack.Workspace = workspaceName
	stack.User = opt.User
	stack.CreateTime = time.Now().Unix()
	//	stack.Templates = make([]string, 0)
	stack.Resources = make(map[string]Resource)

	be := backend.NewBackendHandler()
	//	err = storer.Create(groupName, workspaceName, appName, stack)
	err = be.CreateResource(backendKind, groupName, workspaceName, appName, stack)
	if err != nil {
		sm.Locker.Unlock()
		return log.DebugPrint(err)
	}
	//等待刷入到内存中,不然会出现etcd创建事件的监听晚于删除事件
	sm.Locker.Unlock()
	for {
		_, err := sm.Get(groupName, workspaceName, appName)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if len(desc) == 0 {
		log.DebugPrint("empty")
		return nil
	} else {
		sm.Locker.Lock()
		defer sm.Locker.Unlock()
		//CleanApp:
		//不能直接用deleteAPP,因为锁的原因新创建的资源还没有更新到内存中
		err := stack.AddResources(desc, false)
		if err != nil {
			for _, v := range stack.Resources {

				//删除已创建好的资源
				err2 := stack.RemoveResource(v.Kind, v.Name, false)
				if err2 != nil {
					log.DebugPrint(err2)
				}
			}
			//删应用
			log.DebugPrint("start to delete resource")
			err2 := be.DeleteResource(backendKind, groupName, workspaceName, stack.Name)
			if err2 != nil && err2 != backend.BackendResourceNotFound {
				return log.DebugPrint(err2)
			}
		}
		return err
	}
}

func (sm *AppMananger) get(groupName, workspaceName, name string) (*App, error) {
	group, ok := sm.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	app, ok := workspace.Apps[name]
	if !ok {
		return nil, ErrResourceNotFound
	}
	return &app, nil

}

func (sm *AppMananger) Get(groupName, workspaceName, name string) (AppInterface, error) {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	return sm.get(groupName, workspaceName, name)
}

func (sm *AppMananger) List(groupName string, opt ListOption) ([]AppInterface, error) {
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
func (sm *AppMananger) deleteApp(groupName, workspaceName, name string, opt DeleteOption) error {

	si, err := sm.get(groupName, workspaceName, name)
	if err != nil {
		return err
	}
	be := backend.NewBackendHandler()
	app := si.Info()

	for _, v := range app.Resources {

		err := si.RemoveResource(v.Kind, v.Name, false)
		if err != nil {
			err2 := be.UpdateResource(backendKind, groupName, workspaceName, v.Name, app)
			if err2 != nil {
				log.DebugPrint("store to app backend fail for %v", err)
			}
			return err
		}
	}
	//删应用
	err = be.DeleteResource(backendKind, groupName, workspaceName, app.Name)
	if err != nil && err != backend.BackendResourceNotFound {
		return log.DebugPrint(err)
	}
	return nil
}

func (sm *AppMananger) DeleteApp(groupName, workspaceName, name string, opt DeleteOption) error {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	return sm.deleteApp(groupName, workspaceName, name, opt)
}
