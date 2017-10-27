package configmap

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	"ufleet-deploy/pkg/resource/util"
	"ufleet-deploy/pkg/sign"

	"github.com/ghodss/yaml"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm         *ConfigMapManager
	Controller resource.ObjectController //ConfigMapController
)

type ConfigMapInterface interface {
	Info() *ConfigMap
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
	Metadata() resource.ObjectMeta
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
	Cluster string `json:"cluster"`
}

func GetConfigMapInterface(obj resource.Object) (ConfigMapInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*ConfigMap)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *ConfigMapManager) Lock() {
	p.locker.Lock()
}
func (p *ConfigMapManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *ConfigMapManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := ConfigMap{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp)
	if err != nil {
		return err
	}
	return nil
}

func (p *ConfigMapManager) fillObjectToManager(meta resource.Object) error {

	cm, ok := meta.(*ConfigMap)
	if !ok {
		return fmt.Errorf("object is not correct type")
	}

	group, ok := rm.Groups[cm.Group]
	if !ok {
		return resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[cm.Workspace]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	_, ok = workspace.ConfigMaps[cm.Name]
	if ok {
		return resource.ErrResourceExists
	}

	workspace.ConfigMaps[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *ConfigMapManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *ConfigMapManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group ConfigMapGroup
	group.Workspaces = make(map[string]ConfigMapWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) AddObjectFromBytes(data []byte) error {
	p.Lock()
	defer p.Unlock()
	var res ConfigMap
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res)
	return err

}

func (p *ConfigMapManager) AddWorkspace(groupName string, workspaceName string) error {
	p.Lock()
	defer p.Unlock()
	g, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	_, ok = g.Workspaces[workspaceName]
	if ok {
		return resource.ErrWorkspaceExists
	}

	var ws ConfigMapWorkspace
	ws.ConfigMaps = make(map[string]ConfigMap)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *ConfigMapManager) DeleteWorkspace(groupName string, workspaceName string) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	_, ok = group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}
	delete(group.Workspaces, workspaceName)
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

func (p *ConfigMapManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	return p.Get(group, workspace, resourceName)
}

//注意这里没锁
func (p *ConfigMapManager) get(groupName, workspaceName, resourceName string) (*ConfigMap, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	configmap, ok := workspace.ConfigMaps[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &configmap, nil
}

func (p *ConfigMapManager) Get(group, workspace, resourceName string) (ConfigMapInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *ConfigMapManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, fmt.Errorf("%v:group/%v,workspace/%v", resource.ErrWorkspaceNotFound, groupName, workspaceName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.ConfigMaps {
		t := workspace.ConfigMaps[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ConfigMapManager) ListGroup(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.ConfigMaps {
			t := v.ConfigMaps[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ConfigMapManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var obj corev1.ConfigMap
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != "ConfigMap" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	obj.ResourceVersion = ""
	obj.Annotations = make(map[string]string)
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp ConfigMap
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
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

	err = ph.Create(workspaceName, &obj)
	if err != nil {
		err2 := be.DeleteResource(backendKind, groupName, workspaceName, cp.Name)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil
}

func (p *ConfigMapManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
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
func (p *ConfigMapManager) DeleteNotLock(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.ConfigMaps, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.ConfigMaps, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ConfigMapManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
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

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	if res.MemoryOnly {

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
	Reason     string            `json:"reason"`
	Data       map[string]string `json:"data"`
	DataString string            `json:"datastring"`
}

func (s *ConfigMap) GetStatus() *Status {

	js := Status{ObjectMeta: s.ObjectMeta}
	js.Data = make(map[string]string)
	js.Comment = s.Comment

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	if js.CreateTime == 0 {
		js.CreateTime = runtime.CreationTimestamp.Unix()
	}

	bc, err := yaml.Marshal(runtime.Data)
	if err != nil {
		js.Reason = err.Error()
		return &js

	}
	js.DataString = string(bc)
	js.Data = runtime.Data

	return &js
}
func (s *ConfigMap) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}

func (s *ConfigMap) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitConfigMapController(be backend.BackendHandler) (resource.ObjectController, error) {
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
