package replicationcontroller

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm *ReplicationControllerManager
	/* = &ReplicationControllerManager{
		Groups: make(map[string]ReplicationControllerGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller ReplicationControllerController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type ReplicationControllerController interface {
	Create(group, workspace string, data []byte, opt CreateOptions) error
	Delete(group, workspace, replicationcontroller string, opt DeleteOption) error
	Get(group, workspace, replicationcontroller string) (ReplicationControllerInterface, error)
	List(group, workspace string) ([]ReplicationControllerInterface, error)
	ListGroup(group string) ([]ReplicationControllerInterface, error)
}

type ReplicationControllerInterface interface {
	Info() *ReplicationController
	GetRuntime() (*Runtime, error)
	GetStatus() (*Status, error)
	//	Runtime()
}

type ReplicationControllerManager struct {
	Groups map[string]ReplicationControllerGroup `json:"groups"`
	locker sync.Mutex
}

type ReplicationControllerGroup struct {
	Workspaces map[string]ReplicationControllerWorkspace `json:"Workspaces"`
}

type ReplicationControllerWorkspace struct {
	ReplicationControllers map[string]ReplicationController `json:"replicationcontrollers"`
}

type Runtime struct {
	ReplicationController *corev1.ReplicationController
	Pods                  []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记ReplicationController相关的K8s资源是否仍然存在
//在ReplicationController构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新ReplicationController的信息
type ReplicationController struct {
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	AppStack   string `json:"app"`
	User       string `json:"user"`
	Cluster    string `json:"cluster"`
	CreateTime int64  `json:"createtime"`
	Template   string `json:"template"`
	memoryOnly bool   //用于判定pod是否由k8s自动创建
}

type GetOptions struct{}
type DeleteOption struct{}
type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/replicationcontroller等主动创建的资源.
	//废弃,直接通过ReplicationControllerManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *ReplicationControllerManager) get(groupName, workspaceName, replicationcontrollerName string) (*ReplicationController, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	replicationcontroller, ok := workspace.ReplicationControllers[replicationcontrollerName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &replicationcontroller, nil
}

func (p *ReplicationControllerManager) Get(group, workspace, replicationcontrollerName string) (ReplicationControllerInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, replicationcontrollerName)
}

func (p *ReplicationControllerManager) ListGroup(groupName string) ([]ReplicationControllerInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]ReplicationControllerInterface, 0)
	for _, v := range group.Workspaces {
		for k := range v.ReplicationControllers {
			t := v.ReplicationControllers[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *ReplicationControllerManager) List(groupName, workspaceName string) ([]ReplicationControllerInterface, error) {

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

	pis := make([]ReplicationControllerInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.ReplicationControllers {
		t := workspace.ReplicationControllers[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ReplicationControllerManager) Create(groupName, workspaceName string, data []byte, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewReplicationControllerHandler(groupName, workspaceName)
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

	var svc corev1.ReplicationController
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "ReplicationController" {
		return log.DebugPrint("must and  offer one rc json/yaml data")
	}

	var cp ReplicationController
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
func (p *ReplicationControllerManager) delete(groupName, workspaceName, replicationcontrollerName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.ReplicationControllers, replicationcontrollerName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ReplicationControllerManager) Delete(group, workspace, replicationcontrollerName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewReplicationControllerHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	replicationcontroller, err := p.get(group, workspace, replicationcontrollerName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if replicationcontroller.memoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, replicationcontrollerName)
		if err != nil {
			return log.DebugPrint(err)
		}
		return nil
		//TODO:ufleet创建的数据
	} else {
		be := backend.NewBackendHandler()
		err := be.DeleteResource(backendKind, group, workspace, replicationcontrollerName)
		if err != nil {
			return log.DebugPrint(err)
		}
		err = ph.Delete(workspace, replicationcontrollerName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
		}
		return nil
	}
}

func (j *ReplicationController) Info() *ReplicationController {
	return j
}

func (j *ReplicationController) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewReplicationControllerHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	replicationcontroller, err := ph.Get(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	pods, err := ph.GetPods(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	var runtime Runtime
	runtime.Pods = pods
	runtime.ReplicationController = replicationcontroller
	return &runtime, nil
}

type Status struct {
	Name        string            `json:"name"`
	User        string            `json:"user"`
	Workspace   string            `json:"workspace"`
	Group       string            `json:"group"`
	Images      []string          `json:"images"`
	Containers  []string          `json:"containers"`
	PodNum      int               `json:"podnum"`
	ClusterIP   string            `json:"clusterip"`
	CreateTime  int64             `json:"createtime"`
	Replicas    int32             `json:"replicas"`
	Labels      map[string]string `json:"labels"`
	Annotatiosn map[string]string `json:"annotations"`
	Selectors   map[string]string `json:"selectors"`
	Reason      string            `json:"reason"`
	//	Pods       []string `json:"pods"`
	PodStatus []pk.Status `json:"podstatus"`
	corev1.ReplicationControllerStatus
}

//不包含PodStatus的信息
func K8sReplicationControllerToReplicationControllerStatus(replicationcontroller *corev1.ReplicationController) *Status {
	js := Status{ReplicationControllerStatus: replicationcontroller.Status}
	js.Name = replicationcontroller.Name
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.CreateTime = replicationcontroller.CreationTimestamp.Unix()
	if replicationcontroller.Spec.Replicas != nil {
		js.Replicas = *replicationcontroller.Spec.Replicas
	}

	js.Labels = make(map[string]string)
	if replicationcontroller.Labels != nil {
		js.Labels = replicationcontroller.Labels
	}

	js.Annotatiosn = make(map[string]string)
	if replicationcontroller.Annotations != nil {
		js.Labels = replicationcontroller.Annotations
	}

	js.Selectors = make(map[string]string)
	if replicationcontroller.Spec.Selector != nil {
		js.Selectors = replicationcontroller.Spec.Selector
	}

	for _, v := range replicationcontroller.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
	}
	return &js

}

func (j *ReplicationController) GetStatus() (*Status, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return nil, err
	}

	js := K8sReplicationControllerToReplicationControllerStatus(runtime.ReplicationController)
	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
	}
	info := j.Info()
	js.User = info.User
	js.Group = info.Group
	js.Workspace = info.Workspace

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)
	}

	return js, nil
}

func InitReplicationControllerController(be backend.BackendHandler) (ReplicationControllerController, error) {
	rm = &ReplicationControllerManager{}
	rm.Groups = make(map[string]ReplicationControllerGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group ReplicationControllerGroup
		group.Workspaces = make(map[string]ReplicationControllerWorkspace)
		for i, j := range v.Workspaces {
			var workspace ReplicationControllerWorkspace
			workspace.ReplicationControllers = make(map[string]ReplicationController)
			for m, n := range j.Resources {
				var replicationcontroller ReplicationController
				err := json.Unmarshal([]byte(n), &replicationcontroller)
				if err != nil {
					return nil, fmt.Errorf("init replicationcontroller manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.ReplicationControllers[m] = replicationcontroller
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
