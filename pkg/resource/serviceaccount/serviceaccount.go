package serviceaccount

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
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, serviceaccount string, opt resource.DeleteOption) error
	Update(group, workspace, resource string, newdata []byte) error
	Get(group, workspace, serviceaccount string) (ServiceAccountInterface, error)
	List(group, workspace string) ([]ServiceAccountInterface, error)
	ListGroup(group string) ([]ServiceAccountInterface, error)
}

type ServiceAccountInterface interface {
	Info() *ServiceAccount
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
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

type Runtime struct {
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
	CreateTime int64  `json:"createtime"`
	Template   string `json:"template"`
	memoryOnly bool
}

//注意这里没锁
func (p *ServiceAccountManager) get(groupName, workspaceName, resourceName string) (*ServiceAccount, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	serviceaccount, ok := workspace.ServiceAccounts[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &serviceaccount, nil
}

func (p *ServiceAccountManager) Get(group, workspace, resourceName string) (ServiceAccountInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
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
func (p *ServiceAccountManager) ListGroup(groupName string) ([]ServiceAccountInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]ServiceAccountInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.ServiceAccounts {
			t := v.ServiceAccounts[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ServiceAccountManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewServiceAccountHandler(groupName, workspaceName)
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

	var svc corev1.ServiceAccount
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "ServiceAccount" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	svc.ResourceVersion = ""

	var cp ServiceAccount
	cp.CreateTime = time.Now().Unix()
	cp.Name = svc.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	if opt.App != nil {
		cp.AppStack = *opt.App
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

	return nil

}

//无锁
func (p *ServiceAccountManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return fmt.Errorf("%v: group\\%v", ErrGroupNotFound, groupName)
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return fmt.Errorf("%v: group\\%v,workspace\\%v", ErrWorkspaceNotFound, groupName, workspaceName)
	}

	delete(workspace.ServiceAccounts, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceAccountManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewServiceAccountHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	serviceaccount, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if serviceaccount.memoryOnly {

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

func (p *ServiceAccountManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.ServiceAccount
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewServiceAccountHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

	return nil
}

func (serviceaccount *ServiceAccount) Info() *ServiceAccount {
	return serviceaccount
}

func (s *ServiceAccount) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewServiceAccountHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{ServiceAccount: svc}, nil
}

func (s *ServiceAccount) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.ServiceAccount)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	prefix := "apiVersion: v1\nkind: ServiceAccount"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

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
