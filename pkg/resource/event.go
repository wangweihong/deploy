package resource

import (
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

				if err := oc.Delete(e.Group, e.Workspace, e.Name); err != nil {
					log.ErrorPrint("%v:  event handler/delete:%v", kind, err)
				}
				return

			case cluster.ActionCreate:
				oc.Lock()
				defer oc.Unlock()
				_, err := oc.GetWithoutLock(e.Group, e.Workspace, e.Name)
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

				err = oc.New(p)
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
