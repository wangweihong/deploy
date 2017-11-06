package cluster

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
)

const (
	kind = "cluster"
)

func Init(clusterHostStr string, currentHost string) {

	clusterHost = clusterHostStr
	hostDomain = currentHost

	gws, err := backend.GetExternalWorkspaceList()
	if err != nil {
		panic(err.Error())
	}

	//不能在创建集群后就立即启动informers,因为一旦启动informers,集群的资源事件就开始触发了.o
	//但集群包含多个工作区,所以导致有些工作区还没有来得及写入到cluster的informerControllre中,
	//这些工作区的资源就被当成不关心的工作区的资源给忽略掉了.
	for g, wss := range gws {
		for _, ws := range wss {
			_, err := globalClusterController.CreateOrUpdateCluster(g, ws, false)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	for _, v := range globalClusterController.clusters {
		//只有引用计数为1,则说明该cluster是新创建的,而不是更新的.才会启动informer,
		err := globalClusterController.startClusterInformers(v.Name)
		if err != nil {
			panic(err.Error())
		}

	}

	log.DebugPrint("start to register workspace noticer ", kind)
	wechan, err := backend.RegisterWorkspaceNoticer(kind)
	if err != nil {
		panic(err.Error())
	}
	handleWorkspaceEvent(wechan)
}

func handleWorkspaceEvent(weChan chan backend.WorkspaceEvent) {
	go func() {
		for {
			we := <-weChan
			log.DebugPrint("catch workspace event:%v", we)
			switch we.Action {
			case "delete":
				err := Controller.DeleteCluster(we.Group, we.Workspace)
				if err != nil {
					log.ErrorPrint("delete cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)
				}
			case "set":
				_, err := Controller.CreateOrUpdateCluster(we.Group, we.Workspace, true)
				if err != nil {
					log.ErrorPrint("create cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)
				}
			}
		}

	}()

}
