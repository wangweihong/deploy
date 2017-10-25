package configmap

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	"ufleet-deploy/pkg/resource/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, configmap string, opt resource.DeleteOption) error
	Get(group, workspace, configmap string) (ConfigMapInterface, error)
	Update(group, workspace, resource string, newdata []byte) error
	List(group, workspace string) ([]ConfigMapInterface, error)
	ListGroup(group string) ([]ConfigMapInterface, error)
}

type ConfigMapInterface interface {
	Info() *ConfigMap
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
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

type Runtime struct {
	*corev1.ConfigMap
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记ConfigMap相关的K8s资源是否仍然存在
//在ConfigMap构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新ConfigMap的信息
type ConfigMap struct {
	resource.ObjectMeta
	Cluster    string `json:"cluster"`
	memoryOnly bool
}

//注意这里没锁
func (p *ConfigMapManager) get(groupName, workspaceName, resourceName string) (*ConfigMap, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	configmap, ok := workspace.ConfigMaps[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &configmap, nil
}

func (p *ConfigMapManager) Get(group, workspace, resourceName string) (ConfigMapInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
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

func (p *ConfigMapManager) ListGroup(groupName string) ([]ConfigMapInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]ConfigMapInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.ConfigMaps {
			t := v.ConfigMaps[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ConfigMapManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewConfigMapHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	exts, err := util.ParseJsonOrYaml(data)
	if err != nil {
		return log.DebugPrint(err)
	}

	if len(exts) != 1 {
		return log.DebugPrint("must  offer  one  resource json/yaml data")
	}

	var svc corev1.ConfigMap
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "ConfigMap" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	svc.ResourceVersion = ""
	var cp ConfigMap
	cp.CreateTime = time.Now().Unix()
	cp.Name = svc.Name
	cp.Comment = opt.Comment
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	if opt.App != nil {
		cp.App = *opt.App
	}
	cp.User = opt.User
	//因为pod创建时,触发informer,所以优先创建etcd
	be := backend.NewBackendHandler()
	err = be.CreateResource(backendKind, groupName, workspaceName, cp.Name, cp)
	if err != nil {
		return log.DebugPrint(err)
	}

	err = ph.Create(workspaceName, &svc)
	if err != nil {
		err2 := be.DeleteResource(backendKind, groupName, workspaceName, cp.Name)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil

}
func (p *ConfigMapManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	var newr corev1.ConfigMap
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewConfigMapHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

	return nil
}

//无锁
func (p *ConfigMapManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.ConfigMaps, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewConfigMapHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	res, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.memoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, resourceName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		be := backend.NewBackendHandler()
		err := be.DeleteResource(backendKind, group, workspace, resourceName)
		if err != nil {
			return log.DebugPrint(err)
		}
		err = ph.Delete(workspace, resourceName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
		}

		if !opt.DontCallApp && res.App != "" {
			go func() {
				var re resource.ResourceEvent
				re.Group = group
				re.Workspace = workspace
				re.Kind = resourceKind
				re.Action = resource.ResourceActionDelete
				re.Resource = res.Name
				re.App = res.App

				resource.ResourceEventChan <- re
			}()
		}
		return nil
	}
}

func (configmap *ConfigMap) Info() *ConfigMap {
	return configmap
}

func (s *ConfigMap) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewConfigMapHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{ConfigMap: svc}, nil
}

func (s *ConfigMap) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.ConfigMap)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: v1\nkind: ConfigMap"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	Reason string            `json:"reason"`
	Data   map[string]string `json:"data"`
}

func (s *ConfigMap) GetStatus() *Status {

	js := Status{ObjectMeta: s.ObjectMeta}
	js.Data = make(map[string]string)

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	if js.CreateTime == 0 {
		js.CreateTime = runtime.CreationTimestamp.Unix()
	}

	js.Data = runtime.Data
	return &js
}
func (s *ConfigMap) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
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
