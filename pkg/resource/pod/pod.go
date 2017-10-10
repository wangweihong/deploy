package pod

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/util"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm *PodManager
	/* = &PodManager{
		Groups: make(map[string]PodGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller PodController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type PodController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, pod string, opt DeleteOption) error
	Get(group, workspace, pod string) (PodInterface, error)
	List(group, workspace string) ([]PodInterface, error)
	ListGroup(group string) ([]PodInterface, error)
}

type PodInterface interface {
	Info() *Pod
	GetRuntime() (*Runtime, error)
	GetStatus() (*Status, error)
	GetTemplate() (string, error)
}

type PodManager struct {
	Groups map[string]PodGroup `json:"groups"`
	locker sync.Mutex
}

type PodGroup struct {
	Workspaces map[string]PodWorkspace `json:"Workspaces"`
}

type PodWorkspace struct {
	Pods map[string]Pod `json:"pods"`
}

type Runtime struct {
	Pod *corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Pod相关的K8s资源是否仍然存在
//在Pod构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Pod的信息
type Pod struct {
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

type GetOptions struct {
}
type DeleteOption struct{}

type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/deployment等主动创建的资源.
	//废弃,直接通过PodManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *PodManager) get(groupName, workspaceName, podName string) (*Pod, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	pod, ok := workspace.Pods[podName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &pod, nil
}

func (p *PodManager) Get(group, workspace, podName string) (PodInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, podName)
}

func (p *PodManager) List(groupName, workspaceName string) ([]PodInterface, error) {

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

	pis := make([]PodInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Pods {
		t := workspace.Pods[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *PodManager) ListGroup(groupName string) ([]PodInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]PodInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Pods {
			t := v.Pods[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *PodManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *PodManager) delete(groupName, workspaceName, podName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Pods, podName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *PodManager) Delete(group, workspace, podName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	pod, err := p.get(group, workspace, podName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if pod.memoryOnly {
		ph, err := cluster.NewPodHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, podName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		return nil
	}
}

func (p *Pod) Info() *Pod {
	return p
}

func (p *Pod) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewPodHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, err
	}

	pod, err := ph.Get(p.Workspace, p.Name, cluster.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &Runtime{Pod: pod}, nil
}

func (p *Pod) GetRuntimeDirect() (*Runtime, error) {
	ph, err := cluster.NewPodHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, err
	}

	pod, err := ph.Get(p.Workspace, p.Name, cluster.GetOptions{Direct: true})
	if err != nil {
		return nil, err
	}

	return &Runtime{Pod: pod}, nil
}

type ContainerStatus struct {
	corev1.ContainerStatus
}

type Status struct {
	Name              string            `json:"name"`
	Phase             string            `json:"phase"`
	IP                string            `json:"ip"`
	HostIP            string            `json:"hostip"`
	StartTime         int64             `json:"starttime"`
	Running           int               `json:"running"`
	Total             int               `json:"total"`
	ContainerStatuses []ContainerStatus `json:"containerstatuses"`
}

func V1PodToPodStatus(pod corev1.Pod) *Status {
	var s Status
	s.Name = pod.Name
	ps := pod.Status
	s.Phase = string(ps.Phase)
	s.IP = ps.PodIP
	s.Total = len(pod.Spec.Containers)
	s.HostIP = ps.HostIP
	if ps.StartTime != nil {
		s.StartTime = ps.StartTime.Unix()
	}

	for _, v := range ps.ContainerStatuses {
		s.ContainerStatuses = append(s.ContainerStatuses, ContainerStatus{v})
		if v.Ready {
			s.Running += 1
		}
	}
	return &s

}

func (p *Pod) GetStatus() (*Status, error) {
	runtime, err := p.GetRuntime()
	if err != nil {
		return nil, err
	}
	s := V1PodToPodStatus(*runtime.Pod)

	return s, nil
}

func (p *Pod) GetTemplate() (string, error) {
	if !p.memoryOnly {
		return p.Template, nil
	} else {
		runtime, err := p.GetRuntimeDirect()
		if err != nil {
			return "", log.DebugPrint(err)
		}
		pod := runtime.Pod
		log.DebugPrint(pod.Kind)
		log.DebugPrint(pod.APIVersion)

		t, err := util.GetYamlTemplateFromObject(runtime.Pod)
		if err != nil {
			return "", log.DebugPrint(err)
		}

		return *t, nil
	}
}

func InitPodController(be backend.BackendHandler) (PodController, error) {
	rm = &PodManager{}
	rm.Groups = make(map[string]PodGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group PodGroup
		group.Workspaces = make(map[string]PodWorkspace)
		for i, j := range v.Workspaces {
			var workspace PodWorkspace
			workspace.Pods = make(map[string]Pod)
			for m, n := range j.Resources {
				var pod Pod
				err := json.Unmarshal([]byte(n), &pod)
				if err != nil {
					return nil, fmt.Errorf("init pod manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Pods[m] = pod
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	log.DebugPrint(rm)
	return rm, nil

}
