package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm *ServiceManager
	/* = &ServiceManager{
		Groups: make(map[string]ServiceGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller ServiceController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type ServiceController interface {
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, service string, opt resource.DeleteOption) error
	Update(group, workspace, resource string, newdata []byte) error
	Get(group, workspace, service string) (ServiceInterface, error)
	List(group, workspace string) ([]ServiceInterface, error)
	ListGroup(group string) ([]ServiceInterface, error)
}

type ServiceInterface interface {
	Info() *Service
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
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
	memoryOnly bool
}

//注意这里没锁
func (p *ServiceManager) get(groupName, workspaceName, resourceName string) (*Service, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	service, ok := workspace.Services[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &service, nil
}

func (p *ServiceManager) Get(group, workspace, resourceName string) (ServiceInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *ServiceManager) List(groupName, workspaceName string) ([]ServiceInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()
	log.DebugPrint(p.Groups)

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, fmt.Errorf("%v:group/%v,workspace/%v", ErrWorkspaceNotFound, groupName, workspaceName)
	}

	pis := make([]ServiceInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Services {
		t := workspace.Services[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ServiceManager) ListGroup(groupName string) ([]ServiceInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]ServiceInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Services {
			t := v.Services[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ServiceManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var svc corev1.Service
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "Service" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	svc.ResourceVersion = ""

	var cp Service
	cp.CreateTime = time.Now().Unix()
	cp.Name = svc.Name
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

//无锁
func (p *ServiceManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Services, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewServiceHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	service, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if service.memoryOnly {

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
		return nil
	}
}

func (p *ServiceManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.Service
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewServiceHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
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
	ExternalIPs []string                     `json:"externalips"`
	Reason      string                       `json:"reason"`
	Selector    map[string]string            `json:"selector"`
	Labels      map[string]string            `json:"labels"`
	Ports       []corev1.ServicePort         `json:"ports"`
	Type        string                       `json:"type"`
	PodStatus   []pk.Status                  `json:"podstatus"`
	Ingress     []corev1.LoadBalancerIngress `json:"ingress"`
}

func (s *Service) GetStatus() *Status {
	js := Status{ObjectMeta: s.ObjectMeta}
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

	return &js
}

func InitServiceController(be backend.BackendHandler) (ServiceController, error) {
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
	log.DebugPrint(rm)
	return rm, nil

}
