package hpa

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
	autoscalingv1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
)

var (
	rm         *HorizontalPodAutoscalerManager
	Controller resource.ObjectController //HorizontalPodAutoscalerController
)

type HorizontalPodAutoscalerInterface interface {
	Info() *HorizontalPodAutoscaler
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	//	ObjectStatus() resource.ObjectStatus
}

type HorizontalPodAutoscalerManager struct {
	Groups map[string]HorizontalPodAutoscalerGroup `json:"groups"`
	locker sync.Mutex
}

type HorizontalPodAutoscalerGroup struct {
	Workspaces map[string]HorizontalPodAutoscalerWorkspace `json:"Workspaces"`
}

type HorizontalPodAutoscalerWorkspace struct {
	HorizontalPodAutoscalers map[string]HorizontalPodAutoscaler `json:"hpas"`
}

type Runtime struct {
	*autoscalingv1.HorizontalPodAutoscaler
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记HorizontalPodAutoscaler相关的K8s资源是否仍然存在
//在HorizontalPodAutoscaler构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新HorizontalPodAutoscaler的信息
type HorizontalPodAutoscaler struct {
	resource.ObjectMeta
}

func GetHorizontalPodAutoscalerInterface(obj resource.Object) (HorizontalPodAutoscalerInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*HorizontalPodAutoscaler)
	if !ok {
		return nil, fmt.Errorf("resource object is not hpa type")
	}

	return ri, nil
}

func (p *HorizontalPodAutoscalerManager) Lock() {
	p.locker.Lock()
}

func (p *HorizontalPodAutoscalerManager) Unlock() {
	p.locker.Unlock()
}

func (p *HorizontalPodAutoscalerManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *HorizontalPodAutoscalerManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := HorizontalPodAutoscaler{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

//force:强制填充.用于更新时
func (p *HorizontalPodAutoscalerManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*HorizontalPodAutoscaler)
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
		_, ok = workspace.HorizontalPodAutoscalers[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.HorizontalPodAutoscalers[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *HorizontalPodAutoscalerManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *HorizontalPodAutoscalerManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group HorizontalPodAutoscalerGroup
	group.Workspaces = make(map[string]HorizontalPodAutoscalerWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *HorizontalPodAutoscalerManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *HorizontalPodAutoscalerManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res HorizontalPodAutoscaler
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *HorizontalPodAutoscalerManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws HorizontalPodAutoscalerWorkspace
	ws.HorizontalPodAutoscalers = make(map[string]HorizontalPodAutoscaler)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewHorizontalPodAutoscalerHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	res, err := ph.List(workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	for _, e := range res {

		var o resource.ObjectMeta
		o.Name = e.Name
		o.MemoryOnly = true
		o.Workspace = workspaceName
		o.Group = groupName
		o.User = "kubernetes"
		o.Kind = resourceKind

		err = p.NewObject(o)
		if err != nil && err != resource.ErrResourceExists {
			return log.ErrorPrint(err)
		}
	}
	return nil

}

func (p *HorizontalPodAutoscalerManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *HorizontalPodAutoscalerManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

func (p *HorizontalPodAutoscalerManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	return p.Get(group, workspace, resourceName)
}

func (p *HorizontalPodAutoscalerManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

//注意这里没锁
func (p *HorizontalPodAutoscalerManager) get(groupName, workspaceName, resourceName string) (*HorizontalPodAutoscaler, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	hpa, ok := workspace.HorizontalPodAutoscalers[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &hpa, nil
}

func (p *HorizontalPodAutoscalerManager) Get(group, workspace, resourceName string) (*HorizontalPodAutoscaler, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *HorizontalPodAutoscalerManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.HorizontalPodAutoscalers {
		t := workspace.HorizontalPodAutoscalers[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *HorizontalPodAutoscalerManager) ListGroupObject(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.HorizontalPodAutoscalers {
			t := v.HorizontalPodAutoscalers[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *HorizontalPodAutoscalerManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewHorizontalPodAutoscalerHandler(groupName, workspaceName)
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

	var obj autoscalingv1.HorizontalPodAutoscaler
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	obj.ResourceVersion = ""
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp HorizontalPodAutoscaler
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Comment = opt.Comment
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	cp.Kind = resourceKind

	cp.App = resource.DefaultAppBelong
	if opt.App != nil {
		cp.App = *opt.App
		obj.Annotations[sign.SignUfleetAppKey] = *opt.App
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

func (p *HorizontalPodAutoscalerManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	var newr autoscalingv1.HorizontalPodAutoscaler
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""
	if !res.MemoryOnly {
		if newr.Annotations == nil {
			newr.Annotations = make(map[string]string)
		}
		newr.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	}

	if res.App != "" {
		newr.Annotations[sign.SignUfleetAppKey] = res.App
	}

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewHorizontalPodAutoscalerHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.MemoryOnly {
		err = ph.Update(workspaceName, &newr)
		if err != nil {
			return log.DebugPrint(err)
		}
		return nil
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

//无锁
func (p *HorizontalPodAutoscalerManager) DeleteNotLock(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.HorizontalPodAutoscalers, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *HorizontalPodAutoscalerManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.HorizontalPodAutoscalers, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *HorizontalPodAutoscalerManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewHorizontalPodAutoscalerHandler(group, workspace)
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

		if !opt.DontCallApp && res.App != resource.DefaultAppBelong {
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

func (hpa *HorizontalPodAutoscaler) Info() *HorizontalPodAutoscaler {
	return hpa
}

func (s *HorizontalPodAutoscaler) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewHorizontalPodAutoscalerHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{HorizontalPodAutoscaler: svc}, nil
}

func (s *HorizontalPodAutoscaler) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.HorizontalPodAutoscaler)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion:autoscaling/v1 \nkind: HorizontalPodAutoscaler"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	Reason string `json:"reason"`
}

func (s *HorizontalPodAutoscaler) ObjectStatus() resource.ObjectStatus {
	return s.GetStatus()
}

func (s *HorizontalPodAutoscaler) GetStatus() *Status {

	js := Status{ObjectMeta: s.ObjectMeta}
	js.Comment = s.Comment

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	if js.CreateTime == 0 {
		js.CreateTime = runtime.CreationTimestamp.Unix()
	}

	return &js
}

func (s *HorizontalPodAutoscaler) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitHorizontalPodAutoscalerController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &HorizontalPodAutoscalerManager{}
	rm.Groups = make(map[string]HorizontalPodAutoscalerGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for k, v := range rs {
		var group HorizontalPodAutoscalerGroup
		group.Workspaces = make(map[string]HorizontalPodAutoscalerWorkspace)
		for i, j := range v.Workspaces {
			var workspace HorizontalPodAutoscalerWorkspace
			workspace.HorizontalPodAutoscalers = make(map[string]HorizontalPodAutoscaler)
			for m, n := range j.Resources {
				var hpa HorizontalPodAutoscaler
				err := json.Unmarshal([]byte(n), &hpa)
				if err != nil {
					return nil, fmt.Errorf("init hpa manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.HorizontalPodAutoscalers[m] = hpa
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
