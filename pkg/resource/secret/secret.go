package secret

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
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, secret string, opt DeleteOption) error
	Get(group, workspace, secret string) (SecretInterface, error)
	List(group, workspace string) ([]SecretInterface, error)
}

type SecretInterface interface {
	Info() *Secret
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

type SecretRuntime struct {
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
	memoryOnly bool
}

type GetOptions struct {
}
type DeleteOption struct{}

type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/deployment等主动创建的资源.
	//废弃,直接通过SecretManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *SecretManager) get(groupName, workspaceName, secretName string) (*Secret, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	secret, ok := workspace.Secrets[secretName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &secret, nil
}

func (p *SecretManager) Get(group, workspace, secretName string) (SecretInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, secretName)
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

func (p *SecretManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *SecretManager) delete(groupName, workspaceName, secretName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Secrets, secretName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *SecretManager) Delete(group, workspace, secretName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	secret, err := p.get(group, workspace, secretName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if secret.memoryOnly {
		ph, err := cluster.NewSecretHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, secretName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (secret *Secret) Info() *Secret {
	return secret
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
