package deployment

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
	rm *DeploymentManager
	/* = &DeploymentManager{
		Groups: make(map[string]DeploymentGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller DeploymentController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type DeploymentController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, deployment string, opt DeleteOption) error
	Get(group, workspace, deployment string) (DeploymentInterface, error)
	List(group, workspace string) ([]DeploymentInterface, error)
}

type DeploymentInterface interface {
	Info() *Deployment
}

type DeploymentManager struct {
	Groups map[string]DeploymentGroup `json:"groups"`
	locker sync.Mutex
}

type DeploymentGroup struct {
	Workspaces map[string]DeploymentWorkspace `json:"Workspaces"`
}

type DeploymentWorkspace struct {
	Deployments map[string]Deployment `json:"deployments"`
}

type DeploymentRuntime struct {
	*extensionsv1beta1.Deployment
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Deployment相关的K8s资源是否仍然存在
//在Deployment构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Deployment的信息
type Deployment struct {
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
	//废弃,直接通过DeploymentManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *DeploymentManager) get(groupName, workspaceName, deploymentName string) (*Deployment, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	deployment, ok := workspace.Deployments[deploymentName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &deployment, nil
}

func (p *DeploymentManager) Get(group, workspace, deploymentName string) (DeploymentInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, deploymentName)
}

func (p *DeploymentManager) List(groupName, workspaceName string) ([]DeploymentInterface, error) {

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

	pis := make([]DeploymentInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Deployments {
		t := workspace.Deployments[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *DeploymentManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *DeploymentManager) delete(groupName, workspaceName, deploymentName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Deployments, deploymentName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DeploymentManager) Delete(group, workspace, deploymentName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	deployment, err := p.get(group, workspace, deploymentName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if deployment.memoryOnly {
		ph, err := cluster.NewDeploymentHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, deploymentName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (deployment *Deployment) Info() *Deployment {
	return deployment
}

func InitDeploymentController(be backend.BackendHandler) (DeploymentController, error) {
	rm = &DeploymentManager{}
	rm.Groups = make(map[string]DeploymentGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group DeploymentGroup
		group.Workspaces = make(map[string]DeploymentWorkspace)
		for i, j := range v.Workspaces {
			var workspace DeploymentWorkspace
			workspace.Deployments = make(map[string]Deployment)
			for m, n := range j.Resources {
				var deployment Deployment
				err := json.Unmarshal([]byte(n), &deployment)
				if err != nil {
					return nil, fmt.Errorf("init deployment manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Deployments[m] = deployment
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
