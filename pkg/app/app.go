package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	"ufleet-deploy/pkg/resource/configmap"
	"ufleet-deploy/pkg/resource/cronjob"
	"ufleet-deploy/pkg/resource/daemonset"
	"ufleet-deploy/pkg/resource/deployment"
	"ufleet-deploy/pkg/resource/endpoint"
	"ufleet-deploy/pkg/resource/ingress"
	"ufleet-deploy/pkg/resource/job"
	"ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/replicaset"
	"ufleet-deploy/pkg/resource/replicationcontroller"
	"ufleet-deploy/pkg/resource/secret"
	"ufleet-deploy/pkg/resource/service"
	"ufleet-deploy/pkg/resource/serviceaccount"
	"ufleet-deploy/pkg/resource/statefulset"
	"ufleet-deploy/pkg/resource/util"
)

var (
	Controller AppController
	sm         *AppMananger
)

type AppController interface {
	NewApp(group, workspace, app string, describe []byte, opt CreateOption) error
	DeleteApp(group, workspace, app string, opt DeleteOption) error
	UpdateApp(group, workspace, app string, opt UpdateOption) error
	Get(group, workspaceName, name string) (AppInterface, error)
	List(group string, opt ListOption) ([]AppInterface, error)
	AddAppResource(group, workspace, app string, describe []byte, opt UpdateOption) error
	RemoveAppResource(group, workspace, app string, kind string, resource string) error
}

type AppInterface interface {
	GetTemplates()
	GetResources()
	Info() App
	GetStatus() Status
}

type AppMananger struct {
	Groups map[string]AppGroup `json:"groups"`
	Locker Locker
}

type AppGroup struct {
	Workspaces map[string]AppWorkspace `json:"workspaces"`
}

type AppWorkspace struct {
	Apps map[string]App `json:"apps"`
}

type App struct {
	Name       string              `json:"name"`
	Group      string              `json:"group"`
	Workspace  string              `json:"workspace"`
	Comment    string              `json:"comment"`
	User       string              `json:"user"`
	CreateTime int64               `json:"createtime"`
	Resources  map[string]Resource `json:"resources"` //key: resourceKind_name
}

type Resource struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
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

type UpdateOption struct {
	//	Type string //添加资源,删除资源,更改自身
	Comment    *string
	NewData    []byte
	RemoveList []Resource //移除列表
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
	stack.Resources = make(map[string]Resource)

	be := backend.NewBackendHandler()
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
		return nil
	} else {
		sm.Locker.Lock()
		defer sm.Locker.Unlock()

		log.DebugPrint("start to add resource")
		//CleanApp:
		//不能直接用deleteAPP,因为锁的原因新创建的资源还没有更新到内存中
		err := stack.addResources(desc, false)
		if err != nil {
			for _, v := range stack.Resources {

				//删除已创建好的资源
				err2 := stack.removeResource(v.Kind, v.Name, false)
				if err2 != nil {
					if err2 != resource.ErrResourceNotFound {
						log.ErrorPrint(err2)
					}
				}
			}
			//删应用
			log.DebugPrint("start to delete resource")
			err2 := be.DeleteResource(backendKind, groupName, workspaceName, stack.Name)
			if err2 != nil && err2 != backend.BackendResourceNotFound {
				log.ErrorPrint(err2)
			}
			return log.DebugPrint(err)
		}

		err = be.UpdateResource(backendKind, groupName, workspaceName, appName, stack)
		if err != nil {
			for _, v := range stack.Resources {

				//删除已创建好的资源
				err2 := stack.removeResource(v.Kind, v.Name, false)
				if err2 != nil {
					log.DebugPrint(err2)
				}
			}
			//删应用
			log.DebugPrint("start to delete resource")
			err2 := be.DeleteResource(backendKind, groupName, workspaceName, stack.Name)
			if err2 != nil && err2 != backend.BackendResourceNotFound {
				return log.ErrorPrint(err2)
			}
			return log.DebugPrint(err)
		}
	}
	return nil

}
func (sm *AppMananger) UpdateApp(groupName, workspaceName, appName string, opt UpdateOption) error {
	return nil
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

func (sm *AppMananger) AddAppResource(groupName, workspaceName, appName string, describe []byte, opt UpdateOption) error {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	app, err := sm.get(groupName, workspaceName, appName)
	if err != nil {
		return err
	}

	err = app.addResources(describe, true)
	if err != nil {
		return err
	}

	return nil
}

func (sm *AppMananger) RemoveAppResource(groupName, workspaceName, appName string, kind string, resource string) error {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()
	app, err := sm.get(groupName, workspaceName, appName)
	if err != nil {
		return err
	}
	err = app.removeResource(kind, resource, true)
	if err != nil {
		return err
	}

	return nil

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

		for k := range workspace.Apps {
			t := workspace.Apps[k]
			sis = append(sis, &t)
		}
		return sis, nil
	}

	for _, v := range group.Workspaces {
		for j := range v.Apps {
			t := v.Apps[j]
			sis = append(sis, &t)
		}
	}
	return sis, nil
}

func (sm *AppMananger) deleteApp(groupName, workspaceName, name string, opt DeleteOption) error {

	si, err := sm.get(groupName, workspaceName, name)
	if err != nil {
		return log.DebugPrint(err)
	}
	be := backend.NewBackendHandler()
	app := si.Info()

	for _, v := range app.Resources {

		log.DebugPrint(v.Name)
		err := si.removeResource(v.Kind, v.Name, false)
		if err != nil {
			err2 := be.UpdateResource(backendKind, groupName, workspaceName, v.Name, app)
			if err2 != nil {
				log.ErrorPrint("store to app backend fail for %v", err)
			}
			return log.DebugPrint(err)
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

func generateResourceKey(kind string, name string) string {
	return kind + "_" + name
}

func getResourceKindName(key string) (string, string) {
	s := strings.Split(key, "_")
	return s[0], s[1]
}

func (s *App) GetTemplates() {
}

func (s *App) GetResources() {}

//TODO: 失败时,清掉已经创建的资源
func (s *App) addResources(desc []byte, flush bool) error {
	appName := s.Name
	groupName := s.Group
	workspaceName := s.Workspace
	be := backend.NewBackendHandler()
	exts, uerr := util.ParseJsonOrYaml(desc)
	if uerr != nil {
		return log.DebugPrint(uerr)
	}
	if len(exts) == 0 {
		return log.DebugPrint("must  offer  resource json/yaml data")
	}

	var e error
	var alreadCreateIndex int
	resControllers := make([]struct {
		Controller resource.ObjectController
		Name       string
		Key        string
	}, 0)

	for k := range exts {
		tmp := struct {
			Kind     string `json:"kind"`
			MetaData struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}{}
		var res Resource

		e = json.Unmarshal(exts[k].Raw, &tmp)
		if e != nil {
			e = log.ErrorPrint("create app "+appName+" fail for %v", e)
			alreadCreateIndex = k - 1

			goto Clean
		}

		if strings.TrimSpace(tmp.Kind) == "" || strings.TrimSpace(tmp.MetaData.Name) == "" {
			e = log.ErrorPrint("create app " + appName + " fail for resource kind or name not set")
			alreadCreateIndex = k - 1
			goto Clean
		}

		res.Name = tmp.MetaData.Name
		res.Kind = tmp.Kind

		key := generateResourceKey(res.Kind, res.Name)

		if _, ok := s.Resources[key]; ok {
			e = log.ErrorPrint(" resource %v has exist in app", key)
			alreadCreateIndex = k - 1
			goto Clean
		}

		var rcud resource.ObjectController
		rcud, e = resource.GetResourceController(res.Kind)
		if e != nil {
			alreadCreateIndex = k - 1
			goto Clean
		}

		rc := struct {
			Controller resource.ObjectController
			Name       string
			Key        string
		}{}
		rc.Controller = rcud
		rc.Name = res.Name
		rc.Key = key
		resControllers = append(resControllers, rc)

		opt := resource.CreateOption{}
		opt.App = &appName
		e = rcud.CreateObject(groupName, workspaceName, exts[k].Raw, opt)
		if e != nil {
			alreadCreateIndex = k - 1
			goto Clean
		}
		s.Resources[key] = res
		alreadCreateIndex = k

	}

	if flush {
		e = be.UpdateResource(backendKind, groupName, workspaceName, appName, s)
		if e != nil {
			goto Clean
		}
	}
	return nil
Clean:
	if alreadCreateIndex == -1 {
		return e
	}

	for i := 0; i < alreadCreateIndex+1; i++ {
		err := resControllers[i].Controller.DeleteObject(groupName, workspaceName, resControllers[i].Name, resource.DeleteOption{})
		if err != nil {
			log.ErrorPrint(err)
		}
		delete(s.Resources, resControllers[i].Key)

	}
	return e
}

func (s *App) Info() App {
	return *s
}

type Status struct {
	Name       string                  `json:"name"`
	Group      string                  `json:"group"`
	Workspace  string                  `json:"workspace"`
	Comment    string                  `json:"comment"`
	User       string                  `json:"user"`
	CreateTime int64                   `json:"createtime"`
	Reason     string                  `json:"reason"`
	Resources  map[string]Resource     `json:"resources"`
	Statues    []resource.ObjectStatus `json:"resourcestatuses"`
}

func (app *App) GetStatus() Status {

	var as Status
	as.Statues = make([]resource.ObjectStatus, 0)
	as.Name = app.Name
	as.Group = app.Group
	as.Workspace = app.Workspace
	as.User = app.User
	as.Comment = app.Comment
	as.CreateTime = app.CreateTime
	as.Resources = make(map[string]Resource)
	if app.Resources != nil {
		as.Resources = app.Resources
	}

	statuses := make([]resource.ObjectStatus, 0)
	for _, v := range app.Resources {
		rcud, err := resource.GetResourceController(v.Kind)
		if err != nil {
			as.Reason = err.Error()
			return as
		}

		res, err := rcud.GetObject(app.Group, app.Workspace, v.Name)
		if err != nil {
			as.Reason = err.Error()
			return as
		}

		os := res.ObjectStatus()
		statuses = append(statuses, os)

		switch os.(type) {
		case *configmap.Status:

		case *cronjob.Status:
		case *daemonset.Status:
		case *deployment.Status:
		case *endpoint.Status:
		case *ingress.Status:
		case *job.Status:
		case *pod.Status:
		case *replicaset.Status:
		case *replicationcontroller.Status:
		case *secret.Status:
		case *service.Status:
		case *serviceaccount.Status:
		case *statefulset.Status:
		default:
		}
	}
	as.Statues = append(as.Statues, statuses...)
	return as
}

//更新etcd的数据
//如果flush==true,则说明是一个删除应用触发的删除资源操作,
//不需要刷新到etcd,触发eventWatcher
func (s *App) removeResource(kind string, name string, flush bool) error {
	key := generateResourceKey(kind, name)
	_, ok := s.Resources[key]
	if !ok {
		log.DebugPrint("resource %v doesn't exist in app %v", key, name)
		return ErrResourceNotFoundInApp
	}

	rcud, err := resource.GetResourceController(kind)
	if err != nil {
		return log.DebugPrint(err)
	}

	opt := resource.DeleteOption{}
	if !flush {
		opt.DontCallApp = true
	}
	err = rcud.DeleteObject(s.Group, s.Workspace, name, opt)
	if err != nil && !resource.IsErrorNotFound(err) {
		return log.DebugPrint(err)
	}
	if flush {
		delete(s.Resources, key)

		be := backend.NewBackendHandler()
		//		err := storer.Update(s.Group, s.Workspace, s.Name, s)
		err := be.UpdateResource(backendKind, s.Group, s.Workspace, s.Name, s)
		if err != nil {
			return log.DebugPrint(err)
		}
	}
	return nil

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
