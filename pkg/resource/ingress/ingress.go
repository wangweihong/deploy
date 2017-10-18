package ingress

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
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, ingress string, opt resource.DeleteOption) error
	Get(group, workspace, ingress string) (IngressInterface, error)
	Update(group, workspace, resource string, newdata []byte) error
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
	CreateTime int64  `json:"createne"`
	Template   string `json:"template"`
	memoryOnly bool
}

//注意这里没锁
func (p *IngressManager) get(groupName, workspaceName, resourceName string) (*Ingress, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	ingress, ok := workspace.Ingresss[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &ingress, nil
}

func (p *IngressManager) Get(group, workspace, resourceName string) (IngressInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
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

func (p *IngressManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewIngressHandler(groupName, workspaceName)
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

	var svc extensionsv1beta1.Ingress
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "Ingress" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	var cp Ingress
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
func (p *IngressManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Ingresss, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *IngressManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewIngressHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	ingress, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if ingress.memoryOnly {

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

func (p *IngressManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.Ingress
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewIngressHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

	return nil
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
