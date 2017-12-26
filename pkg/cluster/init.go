package cluster

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
)

const (
	kind = "cluster"
)

func Init(clusterHostStr string, currentHost string) error {

	clusterHost = clusterHostStr
	hostDomain = currentHost

	gws, err := backend.GetExternalWorkspaceList()
	if err != nil {
		return log.DebugPrint(err)
	}
	log.DebugPrint(gws)

	//不能在创建集群后就立即启动informers,因为一旦启动informers,集群的资源事件就开始触发了.o
	//但集群包含多个工作区,所以导致有些工作区还没有来得及写入到cluster的informerControllre中,
	//这些工作区的资源就被当成不关心的工作区的资源给忽略掉了.
	for g, wss := range gws {
		for _, ws := range wss {
			_, _, err := globalClusterController.CreateOrUpdateCluster(g, ws, false)
			if err != nil {
				return log.DebugPrint(err)
			}
		}
	}

	for _, v := range globalClusterController.clusters {
		//只有引用计数为1,则说明该cluster是新创建的,而不是更新的.才会启动informer,
		if v.IllCaused == nil {
			err := globalClusterController.startClusterInformers(v.Name)
			if err != nil {
				return log.DebugPrint(err)
			}
		} else {
			log.ErrorPrint("cluster %v is unhealthy : %v", v.Name, v.IllCaused)
		}
	}

	log.DebugPrint("start to register workspace noticer ", kind)
	wechan, err := backend.RegisterWorkspaceNoticer(kind)
	if err != nil {
		return log.DebugPrint(err)
	}
	log.DebugPrint("start to handle Workspace event")
	go handleWorkspaceEvent(wechan)
	return nil
}

func handleWorkspaceEvent(weChan chan backend.WorkspaceEvent) {
	for {
		we := <-weChan
		go func(we backend.WorkspaceEvent) {
			log.DebugPrint("catch workspace event:%v", we)
			switch we.Action {
			case "delete":
				err := globalClusterController.DeleteCluster(we.Group, we.Workspace)
				if err != nil {
					log.ErrorPrint("delete cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)
					return
				}
				log.DebugPrint("delete cluster(group:%v,workspace:%v)  success", we.Group, we.Workspace)
			case "set":
				cluster, update, err := globalClusterController.CreateOrUpdateCluster(we.Group, we.Workspace, true)
				if err != nil {
					log.ErrorPrint("create cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)
					return
				}
				//创建
				if !update {
					err := cluster.StartInformers()
					if err != nil {
						log.ErrorPrint("cluster(group:%v,workspace:%v)  fail to start informers: %v", we.Group, we.Workspace, err)
						return
					}

					go cluster.HandleHealthyEvent()
				}
				//集群健康
				/*
					if c.IllCaused == nil && !c.informerStart {
						err = c.StartInformers()
						if err != nil {
							log.ErrorPrint("create cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)

						}
					}
				*/

			}
		}(we)
	}

}
