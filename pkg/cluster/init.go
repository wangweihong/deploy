package cluster

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
)

const (
	kind = "cluster"
)

func Init(clusterHostStr string) {

	clusterHost = clusterHostStr

	gws, err := backend.GetExternalWorkspaceList()
	if err != nil {
		panic(err.Error())
	}

	for g, wss := range gws {
		for _, ws := range wss {
			err := Controller.CreateCluster(g, ws)
			if err != nil {
				panic(err.Error())
			}
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
				err := Controller.CreateCluster(we.Group, we.Workspace)
				if err != nil {
					log.ErrorPrint("create cluster(group:%v,workspace:%v)  fail for %v", we.Group, we.Workspace, err)
				}
			}
		}

	}()

}
