package resource

import (
	"fmt"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
)

const (
	clusterConfigMapCreater = "kubernetes"
	eGroup                  = "group"
	eWorkspace              = "workspace"
	eResource               = "resource"
)

func HandleEventWatchFromK8sCluster(echan chan cluster.Event, kind string, oc ObjectController) {
	//可以考虑ufleet创建的configmap直接绑定k8s configmap,
	//在k8s configmap更新的时候再更新绑定的k8s  configmap.
	//在删除的,标记底层资源已经被移除.

	for {
		pe := <-echan
		log.DebugPrint("%v: recieve cluster event : %v", kind, pe)
		if pe.FromUfleet {
			log.DebugPrint("%v:  event %v from ufleet,ignore ", kind, pe)
			return
		}

		go func(e cluster.Event) {
			switch e.Action {
			case cluster.ActionDelete:
				//忽略ufleet主动创建的资源

				if err := oc.Delete(e.Group, e.Workspace, e.Name, DeleteOption{}); err != nil {
					log.ErrorPrint("%v:  event handler/delete:%v", kind, err)
				}
				return

			case cluster.ActionCreate:
				oc.Lock()
				defer oc.Unlock()
				_, err := oc.GetObjectWithoutLock(e.Group, e.Workspace, e.Name)
				if err != nil {
					if err != ErrResourceNotFound {
						log.DebugPrint("%v: event handler create fail:%v", kind, err)
						return
					}
				} else {
					log.DebugPrint("%v: event handler create fail for %v exists", kind, e.Name)
					return
				}

				var p ObjectMeta
				p.Name = e.Name
				p.MemoryOnly = true
				p.Workspace = e.Workspace
				p.Group = e.Group
				p.User = clusterConfigMapCreater

				err = oc.NewObject(p)
				if err != nil {
					log.DebugPrint("%v: event handler create fail for %v", kind, err)
					return
				}

				return

			case cluster.ActionUpdate:

			}
		}(pe)

	}
}

func EtcdEventHandler(e backend.ResourceEvent, cm ObjectController) {

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
			cm.DeleteGroup(e.Group)
		case eWorkspace:
			err := cm.DeleteWorkspace(e.Group, *e.Workspace)
			if err == ErrGroupNotFound {
				log.ErrorPrint("group %v not found", e.Group)
			}
			return

		case eResource:

			//这是一个app事件
			if e.Resource != nil {
				err := cm.Delete(e.Group, *e.Workspace, *e.Resource, DeleteOption{})
				if err != nil {
					log.ErrorPrint("handle delete event(%v) fail for %v", e, err)
				}
			}
			return
		}
	case backend.ActionAdd, backend.ActionCreate, backend.ActionUpdate:
		switch etype {
		case eGroup:
			cm.AddGroup(e.Group)
			return
		case eWorkspace:
			err := cm.AddWorkspace(e.Group, *e.Workspace)
			if err == ErrGroupNotFound {
				log.ErrorPrint(fmt.Sprintf("configmap: group %v doesn't exist in appManager", e.Group))
			}

			return
		case eResource:
			//这是一个资源事件
			err := cm.AddObjectFromBytes([]byte(e.Value))
			if err != nil {
				log.DebugPrint(err)
			}
			return
			//这是一个工作区事件
		}

	default:
		log.ErrorPrint("configmap: app watcher:ingore invalid action:", e.Action)
		return
	}
}
