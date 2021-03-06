package resource

import (
	"strings"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
)

const (
	ClusterObjectCreater = "kubernetes"
	eGroup               = "group"
	eWorkspace           = "workspace"
	eResource            = "resource"
)

func HandleEventWatchFromK8sCluster(echan chan cluster.Event, kind string, oc ObjectController) {
	//可以考虑ufleet创建的configmap直接绑定k8s configmap,
	//在k8s configmap更新的时候再更新绑定的k8s  configmap.
	//在删除的,标记底层资源已经被移除.
	log.DebugPrint(" %v cluster  event handler start !", strings.ToUpper(kind))
	defer log.ErrorPrint("%v cluster  event handler finish !", strings.ToUpper(kind))

	for {
		pe := <-echan

		go func(e cluster.Event) {
			/*
				if e.FromUfleet {
					return
				}
			*/

			switch e.Action {
			case cluster.ActionDelete:
				//清除内存中的数据即可
				obj, err := oc.GetObject(e.Group, e.Workspace, e.Name)
				if err != nil {
					if err == ErrResourceNotFound {
						return
					}
					log.ErrorPrint("%v:  event handler delete 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
					return
				}

				//如果memoryOnly不为true,内存中的数据由etcd去清理
				if obj.Metadata().MemoryOnly {
					err := oc.DeleteObject(e.Group, e.Workspace, e.Name, DeleteOption{MemoryOnly: true})
					if err != nil {
						if err != ErrGroupNotFound || err != ErrWorkspaceNotFound {
							log.ErrorPrint("%v:  event handler delete 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
						}
					}
				}
				return

			case cluster.ActionCreate:
				if e.FromUfleet {
					return
				}
				oc.Lock()
				defer oc.Unlock()

				_, err := oc.GetObjectWithoutLock(e.Group, e.Workspace, e.Name)
				if err != nil {
					if err != ErrResourceNotFound {
						log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
						return
					}
				} else {
					log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v': exists ", kind, e.Group, e.Workspace, e.Name)
					return
				}

				var p ObjectMeta
				p.Name = e.Name
				p.MemoryOnly = true
				p.Workspace = e.Workspace
				p.Group = e.Group
				p.User = ClusterObjectCreater
				p.Kind = kind

				err = oc.NewObject(p)
				if err != nil {
					if err != ErrResourceExists {
						log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)

						return
					}
				}

				return
			case cluster.ActionUpdate:

				return
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

	if etype == eResource {
		log.DebugPrint("%v: handle etcd event:'group:%v workspace:%v resource:%v action:%v'", cm.Kind(), e.Group, *e.Workspace, *e.Resource, e.Action)
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
				log.ErrorPrint("%v: handle etcd event:group %v not found", cm.Kind(), e.Group)
			}
			return

		case eResource:

			if e.Resource != nil {
				//清除内存中的数据即可
				err := cm.DeleteObject(e.Group, *e.Workspace, *e.Resource, DeleteOption{MemoryOnly: true})
				if err != nil {
					log.ErrorPrint("%v: handle etcd event:delete event(%v) fail for %v", e, err)
				}
			}
			return
		}
	case backend.ActionAdd:
		switch etype {
		case eGroup:
			cm.AddGroup(e.Group)
			return
		case eWorkspace:
			err := cm.AddWorkspace(e.Group, *e.Workspace)
			if err == ErrGroupNotFound {
				log.ErrorPrint("%v: handle etcd event: group %v doesn't exist in objectManager", cm.Kind(), e.Group)
			}

			return
		case eResource:
			err := cm.AddObjectFromBytes([]byte(e.Value), true)
			if err != nil {
				log.ErrorPrint("%v: handle etcd event failed to add object for bytes: %v", cm.Kind(), err)
			}
			return
		}

	default:
		log.ErrorPrint("%v: app watcher:ingore invalid action: %v", cm.Kind(), e.Action)
		return
	}

}
