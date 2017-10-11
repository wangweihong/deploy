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

	//	etcdGroupKey          = etcdUfleetKey + "/" + ResourceGroups
	//	etcdWorkspaceKey      = etcdUfleetKey + "/" + ResourceWorkspaces
	etcdAppKey                   = etcdUfleetKey + "/" + ResourceApps
	etcdPodKey                   = etcdUfleetKey + "/" + ResourcePods
	etcdServiceKey               = etcdUfleetKey + "/" + ResourceServices
	etcdSecretKey                = etcdUfleetKey + "/" + ResourceSecrets
	etcdConfigMapKey             = etcdUfleetKey + "/" + ResourceConfigMaps
	etcdEndpointKey              = etcdUfleetKey + "/" + ResourceEndpoints
	etcdServiceAccountKey        = etcdUfleetKey + "/" + ResourceServiceAccounts
	etcdDeploymentKey            = etcdUfleetKey + "/" + ResourceDeployments
	etcdDaemonSetKey             = etcdUfleetKey + "/" + ResourceDaemonSets
	etcdIngressKey               = etcdUfleetKey + "/" + ResourceIngresss
	etcdStatefulSetKey           = etcdUfleetKey + "/" + ResourceStatefulSets
	etcdCronJobKey               = etcdUfleetKey + "/" + ResourceCronJobs
	etcdJobKey                   = etcdUfleetKey + "/" + ResourceJobs
	etcdVolumeKey                = etcdUfleetKey + "/" + ResourceVolumes
	etcdReplicationControllerKey = etcdUfleetKey + "/" + ResourceReplicationControllers

	//	ResourceGroups          = "groups"
	//	ResourceWorkspaces      = "workspaces"
	ResourceApps                   = "apps"
	ResourcePods                   = "pods"
	ResourceServices               = "services"
	ResourceSecrets                = "secrets"
	ResourceConfigMaps             = "configMaps"
	ResourceEndpoints              = "endpoints"
	ResourceServiceAccounts        = "serviceaccounts"
	ResourceDeployments            = "deployments"
	ResourceDaemonSets             = "daemonsets"
	ResourceIngresss               = "ingresss"
	ResourceStatefulSets           = "statefulset"
	ResourceJobs                   = "jobs"
	ResourceCronJobs               = "cronjobs"
	ResourceVolumes                = "volumes"
	ResourceReplicationControllers = "replicationcontrollers"

	ActionDelete = "delete"
	ActionAdd    = "set"
	ActionCreate = "create"
	ActionUpdate = "update"
)

var (
	resources = []string{
		ResourceApps,
		ResourcePods,
		ResourceServices,
		ResourceSecrets,
		ResourceConfigMaps,
		ResourceServiceAccounts,
		ResourceEndpoints,
		ResourceDeployments,
		ResourceDaemonSets,
		ResourceIngresss,
		ResourceStatefulSets,
		ResourceJobs,
		ResourceCronJobs,
		ResourceReplicationControllers,
		//		ResourceGroups,
		//		ResourceWorkspaces,
		//	ResourceVolumes,
	}
	eventParseFail   = fmt.Errorf("can not parse event")
	eventKindInvalid = fmt.Errorf("invalid event kind")

	resourceToBackendkey = map[string]string{
		//		ResourceGroups:          etcdGroupKey,
		//		ResourceWorkspaces:      etcdWorkspaceKey,
		ResourceApps:                   etcdAppKey,
		ResourcePods:                   etcdPodKey,
		ResourceServices:               etcdServiceKey,
		ResourceSecrets:                etcdSecretKey,
		ResourceConfigMaps:             etcdConfigMapKey,
		ResourceEndpoints:              etcdEndpointKey,
		ResourceServiceAccounts:        etcdServiceAccountKey,
		ResourceDeployments:            etcdDeploymentKey,
		ResourceDaemonSets:             etcdDaemonSetKey,
		ResourceIngresss:               etcdIngressKey,
		ResourceStatefulSets:           etcdStatefulSetKey,
		ResourceCronJobs:               etcdCronJobKey,
		ResourceJobs:                   etcdJobKey,
		ResourceVolumes:                etcdVolumeKey,
		ResourceReplicationControllers: etcdReplicationControllerKey,
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
