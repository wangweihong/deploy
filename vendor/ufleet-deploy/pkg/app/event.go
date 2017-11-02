package app

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
)

const (
	eGroup     = "group"
	eWorkspace = "workspace"
	eResource  = "resource"
)

//获得后端监听的事件
//如果组被删除,说明组其中的工作区/资源,以及相应k8s资源都已经被删除,直接删除组即可
//如果工作区删除,说明其中的资源对应的k8s资源已经被清除.直接删除即可
func (a *AppMananger) HandleEvent(e backend.ResourceEvent) {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()
	var etype string
	if e.Workspace == nil {
		etype = eGroup
	} else {
		if e.Resource != nil {
			etype = eResource
		} else {
			etype = eWorkspace
		}
	}

	switch e.Action {
	case backend.ActionDelete:
		//这是一个组事件
		switch etype {
		case eGroup:
			delete(sm.Groups, e.Group)
			return
		case eWorkspace:
			group, ok := sm.Groups[e.Group]
			if !ok {
				delete(group.Workspaces, *e.Workspace)
				sm.Groups[e.Group] = group
				return
			}
		case eResource:
			group, ok := sm.Groups[e.Group]
			if !ok {
				log.ErrorPrint("group %v not found ", e.Group)
				return
			}

			//这是一个app事件
			if e.Resource != nil {
				workspace, ok := group.Workspaces[*e.Workspace]
				if !ok {
					log.ErrorPrint("workspace %v not found", e.Workspace)
					return
				}
				delete(workspace.Apps, *e.Resource)
				group.Workspaces[*e.Workspace] = workspace
				sm.Groups[e.Group] = group
				return
			}
		}
	case backend.ActionAdd, backend.ActionCreate, backend.ActionUpdate:
		switch etype {
		case eGroup:

			if _, ok := sm.Groups[e.Group]; !ok {
				var group AppGroup
				group.Workspaces = make(map[string]AppWorkspace)
				sm.Groups[e.Group] = group
			}
			return
		case eWorkspace:

			group, ok := sm.Groups[e.Group]
			if !ok {
				log.ErrorPrint(fmt.Sprintf("group %v doesn't exist in appManager", e.Group))
				return
			}
			if _, ok := group.Workspaces[*e.Workspace]; !ok {
				var ws AppWorkspace
				ws.Apps = make(map[string]App)
				group.Workspaces[*e.Workspace] = ws
			}
			sm.Groups[e.Group] = group
			return
		case eResource:
			//这是一个资源事件
			var app App
			err := json.Unmarshal([]byte(e.Value), &app)
			if err != nil {
				//error
				log.ErrorPrint("unable to Unmarshal app data \"%v\" for %v", e.Value, err)
				return
			}
			group, ok := sm.Groups[e.Group]
			if !ok {
				//error
				log.ErrorPrint(fmt.Sprintf("group %v doesn't exist in appManager", e.Group))
				return
			}

			ws, ok := group.Workspaces[*e.Workspace]
			if !ok {
				log.ErrorPrint("workspace %v doesn't exist in appManager", *e.Workspace)
				return
			}

			ws.Apps[*e.Resource] = app
			group.Workspaces[*e.Workspace] = ws
			sm.Groups[e.Group] = group
			//这是一个工作区事件
		}

	default:
		log.ErrorPrint("app watcher:ingore invalid action:", e.Action)
		return
	}
}

func ResourceEventHandler() {
	for {
		e := <-resource.ResourceEventChan
		go rehandler(e)
	}

}

func rehandler(e resource.ResourceEvent) {
	sm.Locker.Lock()
	defer sm.Locker.Unlock()

	app, err := sm.get(e.Group, e.Workspace, e.App)
	if err != nil {
		log.ErrorPrint(err)
		return
	}

	err = app.removeResource(e.Kind, e.Resource, true)
	if err != nil {
		if err != ErrResourceNotFoundInApp {
			log.ErrorPrint(err)
		}
		return
	}

	return
}
