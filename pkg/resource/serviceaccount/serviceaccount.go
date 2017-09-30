package serviceaccount

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
	rm *ServiceAccountManager
	/* = &ServiceAccountManager{
		Groups: make(map[string]ServiceAccountGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller ServiceAccountController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type ServiceAccountController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, serviceaccount string, opt DeleteOption) error
	Get(group, workspace, serviceaccount string) (ServiceAccountInterface, error)
	List(group, workspace string) ([]ServiceAccountInterface, error)
}

type ServiceAccountInterface interface {
	Info() *ServiceAccount
}

type ServiceAccountManager struct {
	Groups map[string]ServiceAccountGroup `json:"groups"`
	locker sync.Mutex
}

type ServiceAccountGroup struct {
	Workspaces map[string]ServiceAccountWorkspace `json:"Workspaces"`
}

type ServiceAccountWorkspace struct {
	ServiceAccounts map[string]ServiceAccount `json:"serviceaccounts"`
}

type ServiceAccountRuntime struct {
	*corev1.ServiceAccount
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记ServiceAccount相关的K8s资源是否仍然存在
//在ServiceAccount构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新ServiceAccount的信息
type ServiceAccount struct {
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
	//废弃,直接通过ServiceAccountManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *ServiceAccountManager) get(groupName, workspaceName, serviceaccountName string) (*ServiceAccount, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	serviceaccount, ok := workspace.ServiceAccounts[serviceaccountName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &serviceaccount, nil
}

func (p *ServiceAccountManager) Get(group, workspace, serviceaccountName string) (ServiceAccountInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, serviceaccountName)
}

func (p *ServiceAccountManager) List(groupName, workspaceName string) ([]ServiceAccountInterface, error) {

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

	pis := make([]ServiceAccountInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.ServiceAccounts {
		t := workspace.ServiceAccounts[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ServiceAccountManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *ServiceAccountManager) delete(groupName, workspaceName, serviceaccountName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return fmt.Errorf("%v: group\\%v", ErrGroupNotFound, groupName)
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return fmt.Errorf("%v: group\\%v,workspace\\%v", ErrWorkspaceNotFound, groupName, workspaceName)
	}

	delete(workspace.ServiceAccounts, serviceaccountName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceAccountManager) Delete(group, workspace, serviceaccountName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	serviceaccount, err := p.get(group, workspace, serviceaccountName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if serviceaccount.memoryOnly {
		ph, err := cluster.NewServiceAccountHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, serviceaccountName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (serviceaccount *ServiceAccount) Info() *ServiceAccount {
	return serviceaccount
}

func InitServiceAccountController(be backend.BackendHandler) (ServiceAccountController, error) {
	rm = &ServiceAccountManager{}
	rm.Groups = make(map[string]ServiceAccountGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group ServiceAccountGroup
		group.Workspaces = make(map[string]ServiceAccountWorkspace)
		for i, j := range v.Workspaces {
			var workspace ServiceAccountWorkspace
			workspace.ServiceAccounts = make(map[string]ServiceAccount)
			for m, n := range j.Resources {
				var serviceaccount ServiceAccount
				err := json.Unmarshal([]byte(n), &serviceaccount)
				if err != nil {
					return nil, fmt.Errorf("init serviceaccount manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.ServiceAccounts[m] = serviceaccount
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
