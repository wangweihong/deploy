package daemonset

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
	rm *DaemonSetManager
	/* = &DaemonSetManager{
		Groups: make(map[string]DaemonSetGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller DaemonSetController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type DaemonSetController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, daemonset string, opt DeleteOption) error
	Get(group, workspace, daemonset string) (DaemonSetInterface, error)
	List(group, workspace string) ([]DaemonSetInterface, error)
}

type DaemonSetInterface interface {
	Info() *DaemonSet
}

type DaemonSetManager struct {
	Groups map[string]DaemonSetGroup `json:"groups"`
	locker sync.Mutex
}

type DaemonSetGroup struct {
	Workspaces map[string]DaemonSetWorkspace `json:"Workspaces"`
}

type DaemonSetWorkspace struct {
	DaemonSets map[string]DaemonSet `json:"daemonsets"`
}

type DaemonSetRuntime struct {
	*extensionsv1beta1.DaemonSet
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记DaemonSet相关的K8s资源是否仍然存在
//在DaemonSet构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新DaemonSet的信息
type DaemonSet struct {
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
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/daemonset等主动创建的资源.
	//废弃,直接通过DaemonSetManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *DaemonSetManager) get(groupName, workspaceName, daemonsetName string) (*DaemonSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	daemonset, ok := workspace.DaemonSets[daemonsetName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &daemonset, nil
}

func (p *DaemonSetManager) Get(group, workspace, daemonsetName string) (DaemonSetInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, daemonsetName)
}

func (p *DaemonSetManager) List(groupName, workspaceName string) ([]DaemonSetInterface, error) {

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

	pis := make([]DaemonSetInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.DaemonSets {
		t := workspace.DaemonSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *DaemonSetManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *DaemonSetManager) delete(groupName, workspaceName, daemonsetName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.DaemonSets, daemonsetName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DaemonSetManager) Delete(group, workspace, daemonsetName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	daemonset, err := p.get(group, workspace, daemonsetName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if daemonset.memoryOnly {
		ph, err := cluster.NewDaemonSetHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, daemonsetName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (daemonset *DaemonSet) Info() *DaemonSet {
	return daemonset
}

func InitDaemonSetController(be backend.BackendHandler) (DaemonSetController, error) {
	rm = &DaemonSetManager{}
	rm.Groups = make(map[string]DaemonSetGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group DaemonSetGroup
		group.Workspaces = make(map[string]DaemonSetWorkspace)
		for i, j := range v.Workspaces {
			var workspace DaemonSetWorkspace
			workspace.DaemonSets = make(map[string]DaemonSet)
			for m, n := range j.Resources {
				var daemonset DaemonSet
				err := json.Unmarshal([]byte(n), &daemonset)
				if err != nil {
					return nil, fmt.Errorf("init daemonset manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.DaemonSets[m] = daemonset
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
