package statefulset

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

var (
	rm         *StatefulSetManager
	Controller resource.ObjectController
)

type StatefulSetInterface interface {
	Info() *StatefulSet
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
}

type StatefulSetManager struct {
	Groups map[string]StatefulSetGroup `json:"groups"`
	locker sync.Mutex
}

type StatefulSetGroup struct {
	Workspaces map[string]StatefulSetWorkspace `json:"Workspaces"`
}

type StatefulSetWorkspace struct {
	StatefulSets map[string]StatefulSet `json:"statefulsets"`
}

type Runtime struct {
	StatefulSet *appv1beta1.StatefulSet
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记StatefulSet相关的K8s资源是否仍然存在
//在StatefulSet构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新StatefulSet的信息
type StatefulSet struct {
	resource.ObjectMeta
}

func GetStatefulSetInterface(obj resource.Object) (StatefulSetInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*StatefulSet)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *StatefulSetManager) Lock() {
	p.locker.Lock()
}
func (p *StatefulSetManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *StatefulSetManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := StatefulSet{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *StatefulSetManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*StatefulSet)
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
	if !force {

		_, ok = workspace.StatefulSets[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}

	}
	workspace.StatefulSets[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *StatefulSetManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *StatefulSetManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group StatefulSetGroup
	group.Workspaces = make(map[string]StatefulSetWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *StatefulSetManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res StatefulSet
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *StatefulSetManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws StatefulSetWorkspace
	ws.StatefulSets = make(map[string]StatefulSet)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *StatefulSetManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *StatefulSetManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *StatefulSetManager) get(groupName, workspaceName, resourceName string) (*StatefulSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	statefulset, ok := workspace.StatefulSets[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &statefulset, nil
}

func (p *StatefulSetManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *StatefulSetManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.StatefulSets {
		t := workspace.StatefulSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *StatefulSetManager) ListGroup(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.StatefulSets {
			t := v.StatefulSets[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}
func (p *StatefulSetManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewStatefulSetHandler(groupName, workspaceName)
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

	var obj appv1beta1.StatefulSet
	obj.Annotations = make(map[string]string)
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != "StatefulSet" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	obj.ResourceVersion = ""
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp StatefulSet
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
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

//无锁
func (p *StatefulSetManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.StatefulSets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *StatefulSetManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewStatefulSetHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	res, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.MemoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, resourceName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
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
func (p *StatefulSetManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr appv1beta1.StatefulSet
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewStatefulSetHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	old := *res
	res.Comment = opt.Comment
	be := backend.NewBackendHandler()
	err = be.UpdateResource(backendKind, res.Group, res.Workspace, res.Name, res)
	if err != nil {
		return log.DebugPrint(err)
	}

	err = ph.Update(workspaceName, &newr)
	if err != nil {
		err2 := be.UpdateResource(backendKind, res.Group, res.Workspace, res.Name, &old)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil
}

func (statefulset *StatefulSet) Info() *StatefulSet {
	return statefulset
}
func (s *StatefulSet) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewStatefulSetHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{StatefulSet: svc}, nil
}

func (s *StatefulSet) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.StatefulSet)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: app/v1beta1\nkind: StatefulSet"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	Reason string `json:"reason"`
}

func (s *StatefulSet) GetStatus() *Status {
	js := Status{ObjectMeta: s.ObjectMeta}
	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}

	if js.CreateTime == 0 {
		js.CreateTime = runtime.StatefulSet.CreationTimestamp.Unix()
	}

	return &js

}

func (s *StatefulSet) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}

func (s *StatefulSet) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitStatefulSetController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &StatefulSetManager{}
	rm.Groups = make(map[string]StatefulSetGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group StatefulSetGroup
		group.Workspaces = make(map[string]StatefulSetWorkspace)
		for i, j := range v.Workspaces {
			var workspace StatefulSetWorkspace
			workspace.StatefulSets = make(map[string]StatefulSet)
			for m, n := range j.Resources {
				var statefulset StatefulSet
				err := json.Unmarshal([]byte(n), &statefulset)
				if err != nil {
					return nil, fmt.Errorf("init statefulset manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.StatefulSets[m] = statefulset
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
