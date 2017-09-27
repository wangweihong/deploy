package backend

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
)

const (
	etcdGroupExternalKey = "/ufleet/group"
)

var (
	externalgroupNoticers = make(map[string]struct{})
	externalgroupLock     = sync.Mutex{}
	externalBE            = NewBackendHandler()
)

type CleanFn func() error

type ExternalGroupEvent struct {
	Action string
	Group  string
}

func TuneResourceGroupAccordingToExternalGroup(be BackendHandler, kind string, cleanFn CleanFn) error {

	externalGroups, err := GetExternalGroupList()
	if err != nil {
		return log.DebugPrint(err)
	}

	resourceGroups, err := be.GetResourceGroupList(kind)
	if err != nil {
		return log.DebugPrint(err)
	}

	//外部组不存在则创建
	for k, _ := range externalGroups {
		if _, ok := resourceGroups[k]; !ok {

			err := be.CreateResourceGroup(kind, k)
			if err != nil {
				if err != BackendResourceAlreadyExists {
					return log.DebugPrint(err)
				}
			}
		}
	}

	//已有组不存在于外部组,则删除
	//TODO:需要清理组中相关k8s数据
	for k, _ := range resourceGroups {
		if _, ok := externalGroups[k]; !ok {
			err := be.DeleteResourceGroup(kind, k)
			if err != nil {
				if err == BackendResourceNotFound {
					continue
				}
				return log.DebugPrint(err)
			}
		}
	}
	return nil
}

func GetExternalGroupList() (map[string]string, error) {

	groups := make(map[string]string, 0)

	resp, err := kv.Store.GetNode(etcdGroupExternalKey)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return groups, nil
		}
		return nil, err
	}

	for _, v := range resp.Node.Nodes {
		group := filepath.Base(v.Key)
		groups[group] = group

	}
	return groups, nil
}

/*
func RegisterExternalGroupNoticer(kind string) error {
	externalgroupLock.Lock()
	defer externalgroupLock.Unlock()

	if _, ok := externalgroupNoticers[kind]; ok {
		return fmt.Errorf("externalgroup Noticer \"%v\" has registered", kind)
	}

	externalgroupNoticers[kind] = struct{}{}
	return nil

}
*/

func watchExternalGroupChange() error {

	log.DebugPrint("externalGroupWatcher start to watch ...")
	wechan, err := kv.Store.WatchNode(etcdGroupExternalKey)
	if err != nil {
		return err
	}
	go func(wechan chan kv.WatcheEvent) {
		for {
			we := <-wechan
			if we.Err != nil {
				log.ErrorPrint("externalgroupWatcher watch error: %v", we.Err)
				continue
			}
			res := we.Resp
			if res.Node.Key == etcdGroupExternalKey {
				continue
			}

			var group string
			s := strings.Split(strings.TrimPrefix(res.Node.Key, etcdGroupExternalKey+"/"), "/")
			if len(s) != 1 {
				continue
			}

			log.DebugPrint("externalgroupWatcher recieve group event:", *we.Resp)
			group = s[0]

			var ge ExternalGroupEvent
			switch res.Action {
			case "delete":
				//忽略根Key的事件
				ge.Group = group
				ge.Action = "delete"

			case "set":
				ge.Group = group
				ge.Action = "set"

			default:
				continue
			}

			log.DebugPrint(fmt.Sprintf("start to process external group event : %v", ge))
			//对于所有注册的noticeChan,并发发送ge事件
			/*
				for k, _ := range externalgroupNoticers {
					go func(kind string, ge ExternalGroupEvent) {
						handleExternalGroupEvent(kind, ge.Group, ge.Action)
					}(k, ge)
				}
			*/
			for _, v := range resources {
				go func(kind string, ge ExternalGroupEvent) {
					handleExternalGroupEvent(kind, ge.Group, ge.Action)
				}(v, ge)
			}
		}
	}(wechan)
	return nil
}

func handleExternalGroupEvent(backendKind string, group string, action string) {
	switch action {
	case "delete":
		//直接清理etcd中组数据,触发APP事件
		err := externalBE.DeleteResourceGroup(backendKind, group)
		if err != nil {
			if err != BackendResourceNotFound {
				log.ErrorPrint("resource %v group delete %v fail for %v", backendKind, group, err)
				return
			}
		}

	case "set":
		//直接设置etcd中组数据,触发APP事件

		err := externalBE.CreateResourceGroup(backendKind, group)
		if err != nil {
			if err == BackendResourceAlreadyExists {
				log.DebugPrint(fmt.Errorf("group %v alreay exist", group))
				return
			}
			log.ErrorPrint("resource %v group create group %v fail for %v", backendKind, group, err)
		}
	}
}
