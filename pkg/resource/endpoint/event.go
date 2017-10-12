package endpoint

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
)

const (
	clusterEndpointCreater = "kubernetes"
	eGroup                 = "group"
	eWorkspace             = "workspace"
	eResource              = "resource"
)

func HandleClusterResourceEvent() {
	//可以考虑ufleet创建的endpoint直接绑定k8s endpoint,
	//在k8s endpoint更新的时候再更新绑定的k8s  endpoint.
	//在删除的,标记底层资源已经被移除.

	for {
		pe := <-cluster.EndpointEventChan
		log.DebugPrint("recieve cluster endpoint event : %v", pe)

		go func(e cluster.Event) {
			switch e.Action {
			case cluster.ActionDelete:
				//忽略ufleet主动创建的资源
				if e.FromUfleet {
					log.DebugPrint("endpoint event %v from ufleet,ignore ", e)
					return
				}

				log.DebugPrint("start to delete")
				rm.locker.Lock()
				defer rm.locker.Unlock()
				err := rm.delete(e.Group, e.Workspace, e.Name)
				if err != nil {
					log.ErrorPrint("cluster endpoint  event handler/delete:", err)
					return
				}

				log.DebugPrint("delete success")
				return

			case cluster.ActionCreate:
				if e.FromUfleet {
					log.DebugPrint("endpoint event %v from ufleet,ignore ", e)
					return
				}

				rm.locker.Lock()
				defer rm.locker.Unlock()

				group, ok := rm.Groups[e.Group]
				if !ok {
					log.ErrorPrint("handle cluster endpoint event fail: group \"%v\" not exist", e.Group)
					return
				}

				workspace, ok := group.Workspaces[e.Workspace]
				if !ok {
					log.ErrorPrint("handle cluster endpoint event fail: workspace \"%v\" not exist", e.Workspace)
					return
				}

				x, ok := workspace.Endpoints[e.Name]
				if ok {
					if x.memoryOnly {
						log.ErrorPrint("handle cluster endpoint create event fail: endpoint\"%v\" has exist", e.Name)
					}
					return
				}

				var p Endpoint
				p.Name = e.Name
				p.memoryOnly = true
				p.Workspace = e.Workspace
				p.Group = e.Group
				p.User = clusterEndpointCreater
				workspace.Endpoints[e.Name] = p

				group.Workspaces[e.Workspace] = workspace
				rm.Groups[e.Group] = group
				return

			case cluster.ActionUpdate:
				if e.FromUfleet {
					log.DebugPrint("endpoint event %v from ufleet,ignore ", e)
					return
				}

			}
		}(pe)

	}
}

func EventHandler(e backend.ResourceEvent) {
	rm.locker.Lock()
	defer rm.locker.Unlock()
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
			delete(rm.Groups, e.Group)
			return
		case eWorkspace:
			group, ok := rm.Groups[e.Group]
			if !ok {
				delete(group.Workspaces, *e.Workspace)
				rm.Groups[e.Group] = group
				return
			}
		case eResource:
			group, ok := rm.Groups[e.Group]
			if !ok {
				log.ErrorPrint("group %v not found ", e.Group)
				return
			}

			if e.Resource != nil {
				workspace, ok := group.Workspaces[*e.Workspace]
				if !ok {
					log.ErrorPrint("workspace %v not found", e.Workspace)
					return
				}
				delete(workspace.Endpoints, *e.Resource)
				group.Workspaces[*e.Workspace] = workspace
				rm.Groups[e.Group] = group
				return
			}
		}
	case backend.ActionAdd, backend.ActionCreate, backend.ActionUpdate:
		switch etype {
		case eGroup:

			if _, ok := rm.Groups[e.Group]; !ok {
				var group EndpointGroup
				group.Workspaces = make(map[string]EndpointWorkspace)
				rm.Groups[e.Group] = group
			}
			return
		case eWorkspace:

			group, ok := rm.Groups[e.Group]
			if !ok {
				log.ErrorPrint(fmt.Sprintf("group %v doesn't exist in appManager", e.Group))
				return
			}
			if _, ok := group.Workspaces[*e.Workspace]; !ok {
				var ws EndpointWorkspace
				ws.Endpoints = make(map[string]Endpoint)
				group.Workspaces[*e.Workspace] = ws
			}
			rm.Groups[e.Group] = group
			return
		case eResource:
			//这是一个资源事件
			var app Endpoint
			err := json.Unmarshal([]byte(e.Value), &app)
			if err != nil {
				//error
				log.ErrorPrint("unable to Unmarshal app data \"%v\" for %v", e.Value, err)
				return
			}
			group, ok := rm.Groups[e.Group]
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

			ws.Endpoints[*e.Resource] = app
			group.Workspaces[*e.Workspace] = ws
			rm.Groups[e.Group] = group
			//这是一个工作区事件
		}

	default:
		log.ErrorPrint("app watcher:ingore invalid action:", e.Action)
		return
	}
}
