package service

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
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"
	"ufleet-deploy/pkg/sign"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm         *ServiceManager
	Controller resource.ObjectController
)

type ServiceInterface interface {
	Info() *Service
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
	GetReferenceObjects() ([]ObjectReference, error)
	GetIngresses() ([]*extensionsv1beta1.Ingress, error)
}

type ServiceManager struct {
	Groups map[string]ServiceGroup `json:"groups"`
	locker sync.Mutex
}

type ServiceGroup struct {
	Workspaces map[string]ServiceWorkspace `json:"Workspaces"`
}

type ServiceWorkspace struct {
	Services map[string]Service `json:"services"`
}

type Runtime struct {
	Service *corev1.Service
	Pods    []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Service相关的K8s资源是否仍然存在
//在Service构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Service的信息
type Service struct {
	resource.ObjectMeta
}

func GetServiceInterface(obj resource.Object) (ServiceInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Service)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *ServiceManager) Lock() {
	p.locker.Lock()
}

func (p *ServiceManager) Unlock() {
	p.locker.Unlock()
}

func (p *ServiceManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *ServiceManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Service{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *ServiceManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Service)
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
		_, ok = workspace.Services[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Services[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *ServiceManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *ServiceManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group ServiceGroup
	group.Workspaces = make(map[string]ServiceWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return gs
}

func (p *ServiceManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Service
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *ServiceManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws ServiceWorkspace
	ws.Services = make(map[string]Service)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewServiceHandler(groupName, workspaceName)
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

func (p *ServiceManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *ServiceManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *ServiceManager) get(groupName, workspaceName, resourceName string) (*Service, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	service, ok := workspace.Services[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &service, nil
}

func (p *ServiceManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}
func (p *ServiceManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *ServiceManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()
	log.DebugPrint(p.Groups)

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
	for k := range workspace.Services {
		t := workspace.Services[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ServiceManager) ListGroupObject(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Services {
			t := v.Services[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ServiceManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewServiceHandler(groupName, workspaceName)
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

	var obj corev1.Service
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
	if opt.App != nil {
		obj.Annotations[sign.SignUfleetAppKey] = *opt.App
	}

	var cp Service
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Kind = resourceKind
	cp.Comment = opt.Comment
	cp.Template = string(data)
	cp.App = resource.DefaultAppBelong
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
func (p *ServiceManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Services, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewServiceHandler(group, workspace)
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

func (p *ServiceManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.Service
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//特别注意: 更新service的操作,resource version一定不能为空!!!
	//	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	runtime, err := res.GetRuntime()
	if err != nil {
		return log.DebugPrint(err)
	}

	newr.ResourceVersion = runtime.Service.ResourceVersion
	if runtime.Service.Spec.ClusterIP != "" {
		newr.Spec.ClusterIP = runtime.Service.Spec.ClusterIP
	}
	if !res.MemoryOnly {
		if newr.Annotations == nil {
			newr.Annotations = make(map[string]string)
		}
		newr.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	}
	if res.App != "" {
		newr.Annotations[sign.SignUfleetAppKey] = res.App
	}

	ph, err := cluster.NewServiceHandler(groupName, workspaceName)
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
		return err
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
func (s *Service) Info() *Service {
	return s
}

func (s *Service) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewServiceHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}

	pods, err := ph.GetPods(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{Service: svc, Pods: pods}, nil
}

func (s *Service) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.Service)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	prefix := "apiVersion: v1\nkind: Service"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	ClusterIP   string                       `json:"clusterip"`
	MasterIP    string                       `json:"masterip"`
	ExternalIPs []string                     `json:"externalips"`
	Reason      string                       `json:"reason"`
	Selector    map[string]string            `json:"selector"`
	Labels      map[string]string            `json:"labels"`
	Ports       []corev1.ServicePort         `json:"ports"`
	Containers  []string                     `json:"containers"`
	Type        string                       `json:"type"`
	PodStatus   []pk.Status                  `json:"podstatus"`
	Ingress     []corev1.LoadBalancerIngress `json:"ingress"`
}

func (s *Service) ObjectStatus() resource.ObjectStatus {
	return s.GetStatus()
}

func (s *Service) GetStatus() *Status {
	js := Status{ObjectMeta: s.ObjectMeta}
	js.Containers = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.ExternalIPs = make([]string, 0)
	js.Labels = make(map[string]string)
	js.Selector = make(map[string]string)
	js.Ports = make([]corev1.ServicePort, 0)
	js.Ingress = make([]corev1.LoadBalancerIngress, 0)

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}

	masterip, err := cluster.Controller.GetClusterIP(s.Group, s.Workspace)
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	js.MasterIP = *masterip

	if js.CreateTime == 0 {
		js.CreateTime = runtime.Service.CreationTimestamp.Unix()
	}
	svc := runtime.Service

	js.ExternalIPs = append(js.ExternalIPs, svc.Spec.ExternalIPs...)
	js.ClusterIP = svc.Spec.ClusterIP
	js.Labels = svc.Labels
	js.Selector = svc.Spec.Selector
	js.Ports = append(js.Ports, svc.Spec.Ports...)
	js.Type = string(svc.Spec.Type)
	js.Ingress = append(js.Ingress, svc.Status.LoadBalancer.Ingress...)

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)

		js.PodStatus = append(js.PodStatus, *ps)
	}
	if len(runtime.Pods) != 0 {
		p := runtime.Pods[0]
		for _, v := range p.Spec.Containers {
			js.Containers = append(js.Containers, v.Name)
		}
	}

	return &js
}

func GetStatus(group, workspace, name string) (*Status, error) {
	obj, err := Controller.GetObject(group, workspace, name)
	if err != nil {
		return nil, err
	}

	si, err := GetServiceInterface(obj)
	if err != nil {
		return nil, err
	}
	status := si.GetStatus()
	if status.Reason != "" {
		return nil, err
	}
	return status, nil
}

func (s *Service) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}

func (s *Service) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

type ObjectReference struct {
	corev1.ObjectReference
	Group     string `group`
	Namespace string `json:"workspace"`
}

func (s *Service) GetReferenceObjects() ([]ObjectReference, error) {
	ph, err := cluster.NewServiceHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	apiors, err := ph.GetReferenceResources(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}

	ors := make([]ObjectReference, 0)
	for _, v := range apiors {
		var or ObjectReference
		or.ObjectReference = v
		or.Namespace = s.Workspace
		or.Group = s.Group
		ors = append(ors, or)

	}
	return ors, nil
}

func (s *Service) GetIngresses() ([]*extensionsv1beta1.Ingress, error) {
	ph, err := cluster.NewServiceHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	ings, err := ph.GetIngresses(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return ings, err

}

func InitServiceController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &ServiceManager{}
	rm.Groups = make(map[string]ServiceGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for k, v := range rs {
		var group ServiceGroup
		group.Workspaces = make(map[string]ServiceWorkspace)
		for i, j := range v.Workspaces {
			var workspace ServiceWorkspace
			workspace.Services = make(map[string]Service)
			for m, n := range j.Resources {
				var service Service
				err := json.Unmarshal([]byte(n), &service)
				if err != nil {
					return nil, fmt.Errorf("init service manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Services[m] = service
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
