package service

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
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, service string, opt DeleteOption) error
	Get(group, workspace, service string) (ServiceInterface, error)
	List(group, workspace string) ([]ServiceInterface, error)
}

type ServiceInterface interface {
	Info() *Service
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

type ServiceRuntime struct {
	*corev1.Service
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Service相关的K8s资源是否仍然存在
//在Service构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Service的信息
type Service struct {
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
	//废弃,直接通过ServiceManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *ServiceManager) get(groupName, workspaceName, serviceName string) (*Service, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	service, ok := workspace.Services[serviceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &service, nil
}

func (p *ServiceManager) Get(group, workspace, serviceName string) (ServiceInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, serviceName)
}

func (p *ServiceManager) List(groupName, workspaceName string) ([]ServiceInterface, error) {

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

	pis := make([]ServiceInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Services {
		t := workspace.Services[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ServiceManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *ServiceManager) delete(groupName, workspaceName, serviceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Services, serviceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceManager) Delete(group, workspace, serviceName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	service, err := p.get(group, workspace, serviceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if service.memoryOnly {
		ph, err := cluster.NewServiceHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, serviceName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (service *Service) Info() *Service {
	return service
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
	return rm, nil

}
