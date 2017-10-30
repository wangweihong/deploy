package serviceaccount

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
)

var (
	rm         *ServiceAccountManager
	Controller resource.ObjectController
)

type ServiceAccountInterface interface {
	Info() *ServiceAccount
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
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
	resource.ObjectMeta
	Cluster string `json:"cluster"`
}

func GetServiceAccountInterface(obj resource.Object) (ServiceAccountInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*ServiceAccount)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *ServiceAccountManager) Lock() {
	p.locker.Lock()
}
func (p *ServiceAccountManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *ServiceAccountManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := ServiceAccount{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *ServiceAccountManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*ServiceAccount)
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
		_, ok = workspace.ServiceAccounts[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.ServiceAccounts[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *ServiceAccountManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *ServiceAccountManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group ServiceAccountGroup
	group.Workspaces = make(map[string]ServiceAccountWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceAccountManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res ServiceAccount
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *ServiceAccountManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws ServiceAccountWorkspace
	ws.ServiceAccounts = make(map[string]ServiceAccount)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *ServiceAccountManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *ServiceAccountManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *ServiceAccountManager) get(groupName, workspaceName, resourceName string) (*ServiceAccount, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	serviceaccount, ok := workspace.ServiceAccounts[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &serviceaccount, nil
}

func (p *ServiceAccountManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *ServiceAccountManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.ServiceAccounts {
		t := workspace.ServiceAccounts[k]
		pis = append(pis, &t)
	}

	return pis, nil
}
func (p *ServiceAccountManager) ListGroup(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.ServiceAccounts {
			t := v.ServiceAccounts[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *ServiceAccountManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var obj corev1.ServiceAccount
	obj.Annotations = make(map[string]string)
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != "ServiceAccount" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	obj.ResourceVersion = ""
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp ServiceAccount
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Comment = opt.Comment
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
func (p *ServiceAccountManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return fmt.Errorf("%v: group\\%v", resource.ErrGroupNotFound, groupName)
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return fmt.Errorf("%v: group\\%v,workspace\\%v", resource.ErrWorkspaceNotFound, groupName, workspaceName)
	}

	delete(workspace.ServiceAccounts, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ServiceAccountManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewServiceAccountHandler(group, workspace)
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

func (p *ServiceAccountManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
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

type Status struct {
	resource.ObjectMeta
	Secrts []string `json:"secrets"`
	Reason string   `json:"reason"`
}

func (s *ServiceAccount) GetStatus() *Status {
	js := Status{ObjectMeta: s.ObjectMeta}
	js.Secrts = make([]string, 0)

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}

	if js.CreateTime == 0 {
		js.CreateTime = runtime.CreationTimestamp.Unix()
	}
	for _, v := range runtime.OwnerReferences {
		js.Secrts = append(js.Secrts, v.Name)
	}
	return &js
}

func (s *ServiceAccount) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}

func (s *ServiceAccount) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitServiceAccountController(be backend.BackendHandler) (resource.ObjectController, error) {
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
