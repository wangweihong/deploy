package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"
)

type AppController interface {
	NewApp(group, workspace, app string, describe []byte, opt CreateOptions) error
	DeleteApp(group, workspace, app string, opt DeleteOptions) error
	Get(group, workspaceName, name string) (AppInterface, error)
	List(group string, opt ListOptions) ([]AppInterface, error)
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

func (sm *AppMananger) NewApp(groupName, workspaceName, appName string, desc []byte, opt CreateOptions) error {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	_, err := sm.get(groupName, workspaceName, appName)
	if err == nil {
		return ErrResourceExists
	}
	//加锁
	//
	var stack App
	stack.Name = appName
	stack.Group = groupName
	stack.Workspace = workspaceName
	stack.Templates = make([]string, 0)
	stack.Resources = make(map[string]Resource)

	be := backend.NewBackendHandler()
	//	err = storer.Create(groupName, workspaceName, appName, stack)
	err = be.CreateResource(backendKind, groupName, workspaceName, appName, stack)
	if err != nil {
		return log.DebugPrint(err)
	}

	if len(desc) == 0 {
		return nil
	} else {
		exts, uerr := util.ParseJsonOrYaml(desc)
		if uerr != nil {
			return log.DebugPrint(uerr)
		}
		if len(exts) == 0 {
			return log.DebugPrint("must  offer  resource json/yaml data")
		}

		var err error
		for _, v := range exts {
			tmp := struct {
				Kind     string `json:"kind"`
				MetaData struct {
					Name string `json:"name"`
				} `json:"metadata"`
			}{}
			var res Resource

			err = json.Unmarshal(v.Raw, &tmp)
			if err != nil {
				err = log.ErrorPrint("create app "+appName+" fail for %v", err)
				goto CleanApp
			} else {
				if strings.TrimSpace(tmp.Kind) == "" || strings.TrimSpace(tmp.MetaData.Name) == "" {
					err = log.ErrorPrint("create app " + appName + " fail for resource kind or name not set")
					goto CleanApp
				}

				res.Name = tmp.MetaData.Name
				res.Kind = tmp.Kind
			}
			key := generateResourceKey("Pod", res.Name)

			if _, ok := stack.Resources[key]; ok {
				err = log.ErrorPrint("duplicate resource")
				goto CleanApp
			}

			switch res.Kind {
			case "Pod":
				opt := pod.CreateOptions{}
				opt.App = &appName
				err = pod.Controller.Create(groupName, workspaceName, v.Raw, opt)
				if err != nil {
					goto CleanApp
				}

				stack.Templates = append(stack.Templates, string(v.Raw))
				stack.Resources[key] = res

				//err = storer.Update(groupName, workspaceName, appName, stack)
				err = be.UpdateResource(backendKind, groupName, workspaceName, appName, stack)
				if err != nil {
					err2 := pod.Controller.Delete(groupName, workspaceName, res.Name, pod.DeleteOption{})
					if err2 != nil {
						log.ErrorPrint(err2)
					}
					goto CleanApp
				}

			default:
				err = fmt.Errorf("rsource kind " + res.Kind + " is not supported")
				goto CleanApp
			}

		}
		return nil
	CleanApp:
		//err2 := storer.Delete(groupName, workspaceName, appName)
		err2 := be.DeleteResource(backendKind, groupName, workspaceName, appName)
		if err2 != nil {
			log.ErrorPrint(err2)
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

func (sm *AppMananger) List(groupName string, opt ListOptions) ([]AppInterface, error) {
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
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	si, err := sm.get(groupName, workspaceName, name)
	if err != nil {
		return err
	}
	be := backend.NewBackendHandler()
	app := si.Info()

	err = be.DeleteResource(backendKind, groupName, workspaceName, name)
	if err != nil {
		return log.DebugPrint(err)
	}

	/*
		for _, v := range app.Resources {
			//通知所有资源移除
		}
	*/

	for _, v := range app.Resources {

		key := generateResourceKey(v.Kind, v.Name)
		err := si.RemoveResource(v.Kind, v.Name, false)
		if err != nil {
			//		err2 := storer.Update(groupName, workspaceName, v.Name, app)
			err2 := be.UpdateResource(backendKind, groupName, workspaceName, v.Name, app)
			if err2 != nil {
				log.DebugPrint("store to app backend fail for %v", err)
			}
			return err
		}
		delete(app.Resources, key)
	}
	//删应用
	if !opt.MemoryOnly {
		//	err := storer.Delete(groupName, workspaceName, app.Name)
		err := be.DeleteResource(backendKind, groupName, workspaceName, app.Name)
		if err != nil {
			return log.DebugPrint(err)
		}
	}

	return nil
}
