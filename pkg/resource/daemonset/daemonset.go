package daemonset

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
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, daemonset string, opt resource.DeleteOption) error
	Get(group, workspace, daemonset string) (DaemonSetInterface, error)
	Update(group, workspace, resource string, newdata []byte) error
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
	CreateTime int64  `json:"createtime"`
	Template   string `json:"template"`
	memoryOnly bool
}

//注意这里没锁
func (p *DaemonSetManager) get(groupName, workspaceName, resourceName string) (*DaemonSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	daemonset, ok := workspace.DaemonSets[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &daemonset, nil
}

func (p *DaemonSetManager) Get(group, workspace, resourceName string) (DaemonSetInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
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

func (p *DaemonSetManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewDaemonSetHandler(groupName, workspaceName)
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

	var svc extensionsv1beta1.DaemonSet
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "DaemonSet" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	var cp DaemonSet
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
func (p *DaemonSetManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.DaemonSets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DaemonSetManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewDaemonSetHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	daemonset, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if daemonset.memoryOnly {

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

func (p *DaemonSetManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.DaemonSet
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewDaemonSetHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

	return nil
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
