package configmap

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm *ConfigMapManager
	/* = &ConfigMapManager{
		Groups: make(map[string]ConfigMapGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller ConfigMapController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type ConfigMapController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, configmap string, opt DeleteOption) error
	Get(group, workspace, configmap string) (ConfigMapInterface, error)
	List(group, workspace string) ([]ConfigMapInterface, error)
}

type ConfigMapInterface interface {
	Info() *ConfigMap
}

type ConfigMapManager struct {
	Groups map[string]ConfigMapGroup `json:"groups"`
	locker sync.Mutex
}

type ConfigMapGroup struct {
	Workspaces map[string]ConfigMapWorkspace `json:"Workspaces"`
}

type ConfigMapWorkspace struct {
	ConfigMaps map[string]ConfigMap `json:"configmaps"`
}

type ConfigMapRuntime struct {
	*corev1.ConfigMap
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记ConfigMap相关的K8s资源是否仍然存在
//在ConfigMap构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新ConfigMap的信息
type ConfigMap struct {
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	AppStack   string `json:"app"`
	User       string `json:"user"`
	Cluster    string `json:"cluster"`
	memoryOnly bool
}

type GetOptions struct {
}
type DeleteOption struct{}

type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/deployment等主动创建的资源.
	//废弃,直接通过ConfigMapManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *ConfigMapManager) get(groupName, workspaceName, configmapName string) (*ConfigMap, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	configmap, ok := workspace.ConfigMaps[configmapName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &configmap, nil
}

func (p *ConfigMapManager) Get(group, workspace, configmapName string) (ConfigMapInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, configmapName)
}

func (p *ConfigMapManager) List(groupName, workspaceName string) ([]ConfigMapInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, fmt.Errorf("%v:group/%v,workspace/%v", ErrWorkspaceNotFound, groupName, workspaceName)
	}

	pis := make([]ConfigMapInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.ConfigMaps {
		t := workspace.ConfigMaps[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ConfigMapManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *ConfigMapManager) delete(groupName, workspaceName, configmapName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.ConfigMaps, configmapName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) Delete(group, workspace, configmapName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	configmap, err := p.get(group, workspace, configmapName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if configmap.memoryOnly {
		ph, err := cluster.NewConfigMapHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, configmapName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (configmap *ConfigMap) Info() *ConfigMap {
	return configmap
}

func InitConfigMapController(be backend.BackendHandler) (ConfigMapController, error) {
	rm = &ConfigMapManager{}
	rm.Groups = make(map[string]ConfigMapGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for k, v := range rs {
		var group ConfigMapGroup
		group.Workspaces = make(map[string]ConfigMapWorkspace)
		for i, j := range v.Workspaces {
			var workspace ConfigMapWorkspace
			workspace.ConfigMaps = make(map[string]ConfigMap)
			for m, n := range j.Resources {
				var configmap ConfigMap
				err := json.Unmarshal([]byte(n), &configmap)
				if err != nil {
					return nil, fmt.Errorf("init configmap manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.ConfigMaps[m] = configmap
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
