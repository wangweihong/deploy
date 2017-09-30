package backend

//backend event watcher
import (
	"fmt"
	"strings"
	"sync"
	"time"
	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
)

const (
	etcdUfleetKey = "/ufleet/deploy/v1"

	etcdAppKey     = etcdUfleetKey + "/" + ResourceApps
	etcdServiceKey = etcdUfleetKey + "/" + ResourceServices
	etcdSecretKey  = etcdUfleetKey + "/" + ResourceSecrets
	etcdTaskKey    = etcdUfleetKey + "/" + ResourceTasks
	etcdJobKey     = etcdUfleetKey + "/" + ResourceJobs
	//	etcdGroupKey          = etcdUfleetKey + "/" + ResourceGroups
	//	etcdWorkspaceKey      = etcdUfleetKey + "/" + ResourceWorkspaces
	etcdConfigMapKey      = etcdUfleetKey + "/" + ResourceConfigMaps
	etcdEndpointKey       = etcdUfleetKey + "/" + ResourceEndpoints
	etcdServiceAccountKey = etcdUfleetKey + "/" + ResourceEndpoints
	etcdVolumeKey         = etcdUfleetKey + "/" + ResourceVolumes
	etcdPodKey            = etcdUfleetKey + "/" + ResourcePods

	ResourcePods     = "pods"
	ResourceApps     = "apps"
	ResourceServices = "services"
	ResourceJobs     = "jobs"
	ResourceTasks    = "tasks"
	//	ResourceGroups          = "groups"
	//	ResourceWorkspaces      = "workspaces"
	ResourceSecrets         = "secrets"
	ResourceConfigMaps      = "configMaps"
	ResourceEndpoints       = "endpoints"
	ResourceServiceAccounts = "serviceAccounts"
	ResourceVolumes         = "volumes"

	ActionDelete = "delete"
	ActionAdd    = "set"
	ActionCreate = "create"
	ActionUpdate = "update"
)

var (
	resources = []string{
		ResourceApps,
		ResourcePods,
		//		ResourceServices,
		//		ResourceJobs,
		//		ResourceTasks,
		//		ResourceGroups,
		//		ResourceWorkspaces,
		//		ResourceSecrets,
		//		ResourceConfigMaps,
		//		ResourceEndpoints,
		//	ResourceServiceAccounts,
		//	ResourceVolumes,
	}
	eventParseFail   = fmt.Errorf("can not parse event")
	eventKindInvalid = fmt.Errorf("invalid event kind")

	resourceToBackendkey = map[string]string{
		ResourceApps:     etcdAppKey,
		ResourceServices: etcdServiceKey,
		ResourceSecrets:  etcdSecretKey,
		ResourceTasks:    etcdTaskKey,
		ResourceJobs:     etcdJobKey,
		//		ResourceGroups:          etcdGroupKey,
		ResourceConfigMaps:      etcdConfigMapKey,
		ResourceEndpoints:       etcdEndpointKey,
		ResourceServiceAccounts: etcdServiceAccountKey,
		ResourceVolumes:         etcdVolumeKey,
		ResourcePods:            etcdPodKey,
		//		ResourceWorkspaces:      etcdWorkspaceKey,
	}
)
var (
	locker   sync.Mutex
	noticers = make(map[string]EventHandler)
)

type EventHandler func(ResourceEvent)

//传递的是处于/ufeet/deploy/<kind>剩余的key,key的value,以及action

//注册通知器
func RegisterEventHandler(kind string, fn EventHandler) {
	locker.Lock()
	defer locker.Unlock()
	noticers[kind] = fn

}

func fetchEvent(eventKey string) (string, string, error) {
	s := strings.TrimPrefix(eventKey, etcdUfleetKey+"/")
	slice := strings.SplitN(s, "/", 2)
	if len(slice) != 2 {
		return "", "", eventParseFail
	}
	kind := slice[0]
	remain := slice[1]

	return kind, remain, nil
}

type ResourceEvent struct {
	Group     string
	Workspace *string
	Resource  *string
	Value     string
	Action    string
}

func watchBackendEvent() error {
	wechan, err := kv.Store.WatchNode(etcdUfleetKey)
	if err != nil {
		return err
	}

	go func() {
		for {
			we := <-wechan
			if we.Err != nil {
				log.ErrorPrint(we.Err)
				time.Sleep(1 * time.Second)
				continue
			}

			res := we.Resp

			//可能是新建etcd
			if res.Node.Key == etcdUfleetKey {
				if res.Action == "delete" {
					panic("All Deploy data removed!")
				} else {
					continue
				}
			}

			kind, remain, err := fetchEvent(res.Node.Key)
			if err != nil {
				log.DebugPrint(err)
				continue
			}

			action := res.Action
			value := res.Node.Value

			noticer, ok := noticers[kind]
			if !ok {
				log.DebugPrint(fmt.Errorf("noticer %v doesn't register", kind))
				continue
			}

			re := getEventFromEtcdKey(remain, value, action)

			go noticer(re)

		}
	}()

	return nil
}

func getEventFromEtcdKey(remain string, value string, action string) ResourceEvent {
	var re ResourceEvent
	s := strings.SplitN(remain, "/", -1)
	switch len(s) {
	case 3:
		re.Group = s[0]
		re.Workspace = &s[1]
		re.Resource = &s[2]
	case 2:
		re.Group = s[0]
		re.Workspace = &s[1]
	case 1:
		re.Group = s[0]
	default:
		log.ErrorPrint("invalid event catch by backend")
	}
	re.Value = value
	re.Action = action
	return re

}

func initRootKey() error {
	_, err := kv.Store.CreateDirNode(etcdUfleetKey)
	if err != nil && err != kv.ErrKeyAlreadyExists {
		return err
	}
	return nil
}
