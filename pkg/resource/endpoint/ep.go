package endpoint

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
	rm *EndpointManager
	/* = &EndpointManager{
		Groups: make(map[string]EndpointGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller EndpointController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type EndpointController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, endpoint string, opt DeleteOption) error
	Get(group, workspace, endpoint string) (EndpointInterface, error)
	List(group, workspace string) ([]EndpointInterface, error)
}

type EndpointInterface interface {
	Info() *Endpoint
}

type EndpointManager struct {
	Groups map[string]EndpointGroup `json:"groups"`
	locker sync.Mutex
}

type EndpointGroup struct {
	Workspaces map[string]EndpointWorkspace `json:"Workspaces"`
}

type EndpointWorkspace struct {
	Endpoints map[string]Endpoint `json:"endpoints"`
}

type EndpointRuntime struct {
	*corev1.Endpoints
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Endpoint相关的K8s资源是否仍然存在
//在Endpoint构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Endpoint的信息
type Endpoint struct {
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
	//废弃,直接通过EndpointManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *EndpointManager) get(groupName, workspaceName, endpointName string) (*Endpoint, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	endpoint, ok := workspace.Endpoints[endpointName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &endpoint, nil
}

func (p *EndpointManager) Get(group, workspace, endpointName string) (EndpointInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, endpointName)
}

func (p *EndpointManager) List(groupName, workspaceName string) ([]EndpointInterface, error) {

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

	pis := make([]EndpointInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Endpoints {
		t := workspace.Endpoints[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *EndpointManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *EndpointManager) delete(groupName, workspaceName, endpointName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Endpoints, endpointName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *EndpointManager) Delete(group, workspace, endpointName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	endpoint, err := p.get(group, workspace, endpointName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if endpoint.memoryOnly {
		ph, err := cluster.NewEndpointHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, endpointName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (endpoint *Endpoint) Info() *Endpoint {
	return endpoint
}

func InitEndpointController(be backend.BackendHandler) (EndpointController, error) {
	rm = &EndpointManager{}
	rm.Groups = make(map[string]EndpointGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group EndpointGroup
		group.Workspaces = make(map[string]EndpointWorkspace)
		for i, j := range v.Workspaces {
			var workspace EndpointWorkspace
			workspace.Endpoints = make(map[string]Endpoint)
			for m, n := range j.Resources {
				var endpoint Endpoint
				err := json.Unmarshal([]byte(n), &endpoint)
				if err != nil {
					return nil, fmt.Errorf("init endpoint manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Endpoints[m] = endpoint
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
