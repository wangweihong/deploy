package backend

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"

	"github.com/astaxie/beego"
)

//
//{"action":"get","node":{"key":"/ufleet/cluster/workspace","dir":true,"nodes":[{"key":"/ufleet/cluster/workspace/namespace","value":"{\"c_cpu_min\": 0.001, \"c_mem_default\": 512, \"pod_mem_max\": 2048, \"group\": \"Group\", \"name\": \"namespace\", \"mem\": 18.336, \"pod_mem_min\": 2, \"c_cpu_max\": 0.6, \"datetime\": \"2017-08-08 09:34:31\", \"cluster_name\": \"JGHSUIGYAIUGHKJ\", \"creater\": \"admin\", \"c_mem_max\": 1024, \"c_mem_min\": 1, \"pod_cpu_min\": 0.001, \"pod_cpu_max\": 0.6, \"c_cpu_default\": 0.1, \"cpu\": 9.6}","modifiedIndex":301,"createdIndex":301}],"modifiedIndex":301,"createdIndex":301}}
const (
	externalWorkspaceKey = "/ufleet/cluster/workspace"
)

var (
	//
	workspaceNoticers = make(map[string]chan WorkspaceEvent)
	workspaceLock     = sync.Mutex{}
	workspaceBE       = NewBackendHandler()
)

type Workspace struct {
	Group     string `json:"group"`
	Workspace string `json:"name"`
}

func RegisterWorkspaceNoticer(kind string) (chan WorkspaceEvent, error) {
	workspaceLock.Lock()
	defer workspaceLock.Unlock()

	_, ok := workspaceNoticers[kind]
	if ok {
		return nil, fmt.Errorf("noticer %v has registered", kind)
	}

	ch := make(chan WorkspaceEvent)
	workspaceNoticers[kind] = ch
	return ch, nil

}

//必须在GroupTuner后面执行
func TuneResourcesWorkspaceAccordingToExternalWorkspace(be BackendHandler, kind string) error {
	gws, err := GetExternalWorkspaceList()
	if err != nil {
		return log.DebugPrint(err)
	}

	//不管存不存在,直接创建
	rGroups, err := be.GetResourceAllGroup(kind)
	//这时候,可能resource key还没有
	if err != nil && err != BackendResourceNotFound {
		//	if err != nil {
		return log.DebugPrint(err)
	}

	for group, wslist := range gws {
		for _, j := range wslist {
			err := be.CreateResourceWorkspace(kind, group, j)
			if err != nil && err != BackendResourceAlreadyExists {
				return err
			}
		}
	}

	for group, v := range rGroups {
		wslit := gws[group]
		for ws, _ := range v.Workspaces {
			var found bool
			for _, ews := range wslit {
				if ews == ws {
					found = true
				}
			}

			if found {
				continue
			} else {
				err := be.DeleteResourceWorkspace(kind, group, ws)
				if err != nil && err != BackendResourceNotFound {
					return log.DebugPrint(err)
				}
			}
		}
	}

	return nil
}

func GetExternalWorkspaceList() (map[string][]string, error) {
	gws := make(map[string][]string)

	resp, err := kv.Store.GetChildNode(externalWorkspaceKey)
	if err != nil && err != kv.ErrKeyNotFound {
		return nil, log.DebugPrint(err)
	}

	if err == kv.ErrKeyNotFound {
		return gws, nil
	}

	for _, v := range resp {
		var w Workspace
		err := json.Unmarshal([]byte(v.Value), &w)
		if err != nil {
			return nil, log.DebugPrint(err)
		}
		ws, ok := gws[w.Group]
		if ok {
			ws = append(ws, w.Workspace)
			gws[w.Group] = ws
		} else {
			ws := make([]string, 0)
			ws = append(ws, w.Workspace)
			gws[w.Group] = ws
		}

	}
	return gws, nil
}

type WorkspaceEvent struct {
	Action    string
	Group     string
	Workspace string
}

func GetGroupWorkspace(groupName string) ([]string, error) {

	ws := make([]string, 0)

	resp, err := kv.Store.GetChildNode(externalWorkspaceKey)
	if err != nil {
		return nil, err
	}

	for _, v := range resp {
		var w Workspace
		err := json.Unmarshal([]byte(v.Value), &w)
		if err != nil {
			return nil, err
		}

		if w.Group == groupName {
			ws = append(ws, w.Workspace)
		}

	}

	return ws, nil

}

func watchWorkspaceChange() error {
	wechan, err := kv.Store.WatchNode(externalWorkspaceKey)
	if err != nil {
		return err
	}

	go func() {
		for {
			we := <-wechan
			if we.Err != nil {
				log.ErrorPrint("watch workspace change for ", we.Err)
				time.Sleep(1 * time.Second)
				continue
			}

			res := we
			if res.Node.Key == externalWorkspaceKey {
				continue
			}

			var action string
			var value string
			switch res.Action {
			case kv.ActionCreate:

				action = res.Action
				value = res.Node.Value

			case kv.ActionDelete:
				action = res.Action
				//				value = res.PrevNode.Value
			}

			var w Workspace
			err := json.Unmarshal([]byte(value), &w)
			if err != nil {
				beego.Error("cannot unmarshal value '%v' for ", value, err)
				continue
			}
			var event WorkspaceEvent
			event.Action = action
			event.Group = w.Group
			event.Workspace = w.Workspace

			log.DebugPrint("recieve workspace event %v/%v/%v", event.Group, event.Workspace, event.Action)
			for _, k := range resources {
				go func(kind string, we WorkspaceEvent) {
					handleWorkspaceEvent(kind, we.Group, we.Workspace, we.Action)
				}(k, event)
			}

			for _, k := range workspaceNoticers {
				go func(c chan WorkspaceEvent, we WorkspaceEvent) {
					c <- we
				}(k, event)
			}

			//需要通知cluster进行相应的处理
		}
	}()

	return nil
}

func handleWorkspaceEvent(backendKind string, group string, workspace string, action string) {
	switch action {
	case "delete":
		//直接清理etcd中组数据,触发APP事件
		err := workspaceBE.DeleteResourceWorkspace(backendKind, group, workspace)
		if err != nil {
			if err != BackendResourceNotFound {
				log.ErrorPrint("resource %v workspace delete %v fail for %v", backendKind, group, err)
				return
			}
		}

	case "set":
		//直接设置etcd中组数据,触发APP事件

		err := workspaceBE.CreateResourceWorkspace(backendKind, group, workspace)
		if err != nil {
			if err == BackendResourceAlreadyExists {
				//				log.DebugPrint(fmt.Errorf("resource %v group %v workspace %v alreay exist", backendKind, group, workspace))
				return
			}
			log.ErrorPrint("resource %v  create group %v workspace %v fail for %v", backendKind, group, workspace, err)
		}
	}
}
