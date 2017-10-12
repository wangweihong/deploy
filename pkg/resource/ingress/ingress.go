package ingress

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"

	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm *IngressManager
	/* = &IngressManager{
		Groups: make(map[string]IngressGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller IngressController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type IngressController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, ingress string, opt DeleteOption) error
	Get(group, workspace, ingress string) (IngressInterface, error)
	List(group, workspace string) ([]IngressInterface, error)
}

type IngressInterface interface {
	Info() *Ingress
}

type IngressManager struct {
	Groups map[string]IngressGroup `json:"groups"`
	locker sync.Mutex
}

type IngressGroup struct {
	Workspaces map[string]IngressWorkspace `json:"Workspaces"`
}

type IngressWorkspace struct {
	Ingresss map[string]Ingress `json:"ingresss"`
}

type IngressRuntime struct {
	*extensionsv1beta1.Ingress
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Ingress相关的K8s资源是否仍然存在
//在Ingress构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Ingress的信息
type Ingress struct {
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
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/ingress等主动创建的资源.
	//废弃,直接通过IngressManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *IngressManager) get(groupName, workspaceName, ingressName string) (*Ingress, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	ingress, ok := workspace.Ingresss[ingressName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &ingress, nil
}

func (p *IngressManager) Get(group, workspace, ingressName string) (IngressInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, ingressName)
}

func (p *IngressManager) List(groupName, workspaceName string) ([]IngressInterface, error) {

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

	pis := make([]IngressInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Ingresss {
		t := workspace.Ingresss[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *IngressManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *IngressManager) delete(groupName, workspaceName, ingressName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Ingresss, ingressName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *IngressManager) Delete(group, workspace, ingressName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ingress, err := p.get(group, workspace, ingressName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if ingress.memoryOnly {
		ph, err := cluster.NewIngressHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, ingressName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (ingress *Ingress) Info() *Ingress {
	return ingress
}

func InitIngressController(be backend.BackendHandler) (IngressController, error) {
	rm = &IngressManager{}
	rm.Groups = make(map[string]IngressGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group IngressGroup
		group.Workspaces = make(map[string]IngressWorkspace)
		for i, j := range v.Workspaces {
			var workspace IngressWorkspace
			workspace.Ingresss = make(map[string]Ingress)
			for m, n := range j.Resources {
				var ingress Ingress
				err := json.Unmarshal([]byte(n), &ingress)
				if err != nil {
					return nil, fmt.Errorf("init ingress manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Ingresss[m] = ingress
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}