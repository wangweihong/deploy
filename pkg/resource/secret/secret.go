package secret

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
	rm *SecretManager
	/* = &SecretManager{
		Groups: make(map[string]SecretGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller SecretController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type SecretController interface {
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, secret string, opt resource.DeleteOption) error
	Get(group, workspace, secret string) (SecretInterface, error)
	Update(group, workspace, resource string, newdata []byte) error
	List(group, workspace string) ([]SecretInterface, error)
	ListGroup(group string) ([]SecretInterface, error)
}

type SecretInterface interface {
	Info() *Secret
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
}

type SecretManager struct {
	Groups map[string]SecretGroup `json:"groups"`
	locker sync.Mutex
}

type SecretGroup struct {
	Workspaces map[string]SecretWorkspace `json:"Workspaces"`
}

type SecretWorkspace struct {
	Secrets map[string]Secret `json:"secrets"`
}

type Runtime struct {
	*corev1.Secret
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Secret相关的K8s资源是否仍然存在
//在Secret构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Secret的信息
type Secret struct {
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	AppStack   string `json:"app"`
	User       string `json:"user"`
	Cluster    string `json:"cluster"`
	Template   string `json:"template"`
	CreateTime int64  `json:"createtime"`
	memoryOnly bool
}

//注意这里没锁
func (p *SecretManager) get(groupName, workspaceName, resourceName string) (*Secret, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	secret, ok := workspace.Secrets[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &secret, nil
}

func (p *SecretManager) Get(group, workspace, resourceName string) (SecretInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *SecretManager) List(groupName, workspaceName string) ([]SecretInterface, error) {

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

	pis := make([]SecretInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Secrets {
		t := workspace.Secrets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *SecretManager) ListGroup(groupName string) ([]SecretInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]SecretInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Secrets {
			t := v.Secrets[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *SecretManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewSecretHandler(groupName, workspaceName)
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

	var svc corev1.Secret
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "Secret" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	var cp Secret
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
}

//无锁
func (p *SecretManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Secrets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *SecretManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	secret, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	ph, err := cluster.NewSecretHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	if secret.memoryOnly {

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

func (p *SecretManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.Secret
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewSecretHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

	return nil
}
func (secret *Secret) Info() *Secret {
	return secret
}

func (s *Secret) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewSecretHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{Secret: svc}, nil
}

func (s *Secret) GetTemplate() (string, error) {
	if !s.memoryOnly {
		return s.Template, nil
	} else {
		runtime, err := s.GetRuntime()
		if err != nil {
			return "", err
		}
		t, err := util.GetYamlTemplateFromObject(runtime.Secret)
		if err != nil {
			return "", log.DebugPrint(err)
		}
		return *t, nil

	}
}

func InitSecretController(be backend.BackendHandler) (SecretController, error) {
	rm = &SecretManager{}
	rm.Groups = make(map[string]SecretGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for k, v := range rs {
		var group SecretGroup
		group.Workspaces = make(map[string]SecretWorkspace)
		for i, j := range v.Workspaces {
			var workspace SecretWorkspace
			workspace.Secrets = make(map[string]Secret)
			for m, n := range j.Resources {
				var secret Secret
				err := json.Unmarshal([]byte(n), &secret)
				if err != nil {
					return nil, fmt.Errorf("init secret manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Secrets[m] = secret
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
