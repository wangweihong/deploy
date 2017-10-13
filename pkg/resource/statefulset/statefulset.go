package statefulset

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

var (
	rm *StatefulSetManager
	/* = &StatefulSetManager{
		Groups: make(map[string]StatefulSetGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller StatefulSetController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type StatefulSetController interface {
	Create(group, workspace string, data []byte, opt CreateOptions) error
	Delete(group, workspace, statefulset string, opt DeleteOption) error
	Get(group, workspace, statefulset string) (StatefulSetInterface, error)
	List(group, workspace string) ([]StatefulSetInterface, error)
}

type StatefulSetInterface interface {
	Info() *StatefulSet
}

type StatefulSetManager struct {
	Groups map[string]StatefulSetGroup `json:"groups"`
	locker sync.Mutex
}

type StatefulSetGroup struct {
	Workspaces map[string]StatefulSetWorkspace `json:"Workspaces"`
}

type StatefulSetWorkspace struct {
	StatefulSets map[string]StatefulSet `json:"statefulsets"`
}

type StatefulSetRuntime struct {
	*appv1beta1.StatefulSet
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记StatefulSet相关的K8s资源是否仍然存在
//在StatefulSet构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新StatefulSet的信息
type StatefulSet struct {
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

type GetOptions struct {
}
type DeleteOption struct{}

type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/statefulset等主动创建的资源.
	//废弃,直接通过StatefulSetManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *StatefulSetManager) get(groupName, workspaceName, resourceName string) (*StatefulSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	statefulset, ok := workspace.StatefulSets[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &statefulset, nil
}

func (p *StatefulSetManager) Get(group, workspace, resourceName string) (StatefulSetInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *StatefulSetManager) List(groupName, workspaceName string) ([]StatefulSetInterface, error) {

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

	pis := make([]StatefulSetInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.StatefulSets {
		t := workspace.StatefulSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *StatefulSetManager) Create(groupName, workspaceName string, data []byte, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewStatefulSetHandler(groupName, workspaceName)
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

	var svc appv1beta1.StatefulSet
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "StatefulSet" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	var cp StatefulSet
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
func (p *StatefulSetManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.StatefulSets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *StatefulSetManager) Delete(group, workspace, resourceName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewStatefulSetHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	statefulset, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if statefulset.memoryOnly {

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

func (statefulset *StatefulSet) Info() *StatefulSet {
	return statefulset
}

func InitStatefulSetController(be backend.BackendHandler) (StatefulSetController, error) {
	rm = &StatefulSetManager{}
	rm.Groups = make(map[string]StatefulSetGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group StatefulSetGroup
		group.Workspaces = make(map[string]StatefulSetWorkspace)
		for i, j := range v.Workspaces {
			var workspace StatefulSetWorkspace
			workspace.StatefulSets = make(map[string]StatefulSet)
			for m, n := range j.Resources {
				var statefulset StatefulSet
				err := json.Unmarshal([]byte(n), &statefulset)
				if err != nil {
					return nil, fmt.Errorf("init statefulset manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.StatefulSets[m] = statefulset
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
