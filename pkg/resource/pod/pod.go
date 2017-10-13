package pod

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/util"
	cadvisor "ufleet-deploy/util/cadvisor"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apis/pkg/api/errors"

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
	Create(group, workspace string, data []byte, opt CreateOptions) error
	Delete(group, workspace, pod string, opt DeleteOption) error
	Get(group, workspace, pod string) (PodInterface, error)
	Update(group, workspace, pod string, newdata []byte) error
	List(group, workspace string) ([]PodInterface, error)
	ListGroup(group string) ([]PodInterface, error)
}

type PodInterface interface {
	Info() *Pod
	GetRuntime() (*Runtime, error)
	GetStatus() (*Status, error)
	GetTemplate() (string, error)
	Log(c string) (string, error)
	Stat(c string) ([]ContainerStat, error)
	Terminal(containerName string) (string, error)
	Event() ([]corev1.Event, error)
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

type GetOptions struct{}
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

func (p *PodManager) Create(groupName, workspaceName string, data []byte, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewPodHandler(groupName, workspaceName)
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

	var pod corev1.Pod
	err = json.Unmarshal(exts[0].Raw, &pod)
	if err != nil {
		return log.DebugPrint(err)
	}

	if pod.Kind != "Pod" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	var cp Pod
	cp.CreateTime = time.Now().Unix()
	cp.Name = pod.Name
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

	err = ph.Create(workspaceName, &pod)
	if err != nil {
		err2 := be.DeleteResource(backendKind, groupName, workspaceName, cp.Name)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil

}

func (p *PodManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	/*
		if r.Info().Template != "" {

		}
	*/
	//log.DebugPrint(string(data))
	var newr corev1.Pod
	/*
		err = json.Unmarshal(data, &newr)
		if err != nil {
			return log.DebugPrint(err)
		}
	*/
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewPodHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	err = ph.Update(workspaceName, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}

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
	ph, err := cluster.NewPodHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	p.locker.Lock()
	defer p.locker.Unlock()
	pod, err := p.get(group, workspace, podName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if pod.memoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, podName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		be := backend.NewBackendHandler()
		err := be.DeleteResource(backendKind, group, workspace, podName)
		if err != nil {
			return log.DebugPrint(err)
		}
		err = ph.Delete(workspace, podName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
		}
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
	Name string `json:"name"`
	//TODO:修正pod的状态
	Phase             string            `json:"phase"`
	IP                string            `json:"ip"`
	HostIP            string            `json:"hostip"`
	StartTime         int64             `json:"starttime"`
	Running           int               `json:"running"`
	Total             int               `json:"total"`
	Restarts          int               `json:"restarts"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	RestartPolicy     string            `json:"restartpolicy"`
	ContainerStatuses []ContainerStatus `json:"containerstatuses"`
	ContainerSpecs    []ContainerSpec   `json:"containerspec"`
}

type PodSpec struct {
	ContainerSpecs []ContainerSpec `json:"containerspecs"`
}

//因为omitempty,会出现部分字段丢失
type ContainerSpec struct {
	Name                     string                          `json:"name"`
	Image                    string                          `json:"image"`
	Command                  []string                        `json:"command"`
	Args                     []string                        `json:"args"`
	WorkingDir               string                          `json:"workingDir"`
	Ports                    []corev1.ContainerPort          `json:"ports"`
	EnvFrom                  []corev1.EnvFromSource          `json:"envFrom"`
	Env                      []corev1.EnvVar                 `json:"env"`
	Resources                corev1.ResourceRequirements     `json:"resources"`
	VolumeMounts             []corev1.VolumeMount            `json:"volumeMounts"`
	LivenessProbe            *corev1.Probe                   `json:"livenessProbe"`
	ReadinessProbe           *corev1.Probe                   `json:"readinessProbe"`
	Lifecycle                *corev1.Lifecycle               `json:"lifecycle"`
	TerminationMessagePath   string                          `json:"terminationMessagePath"`
	TerminationMessagePolicy corev1.TerminationMessagePolicy `json:"terminationMessagePolicy"`
	ImagePullPolicy          corev1.PullPolicy               `json:"imagePullPolicy"`
	SecurityContext          *corev1.SecurityContext         `json:"securityContext"`
	StdinOnce                bool                            `json:"stdinOnce"`
	Stdin                    bool                            `json:"stdin"`
	TTY                      bool                            `json:"tty"`
}

func k8sContainerSpecTran(cspec *corev1.Container) *ContainerSpec {
	cs := new(ContainerSpec)
	cs.Command = make([]string, 0)
	cs.Args = make([]string, 0)
	cs.Ports = make([]corev1.ContainerPort, 0)
	cs.EnvFrom = make([]corev1.EnvFromSource, 0)
	cs.Env = make([]corev1.EnvVar, 0)
	cs.VolumeMounts = make([]corev1.VolumeMount, 0)

	cs.Name = cspec.Name
	cs.Image = cspec.Image
	if len(cspec.Command) != 0 {
		cs.Command = cspec.Command
	}
	if len(cspec.Args) != 0 {
		cs.Args = cspec.Args
	}
	cs.WorkingDir = cspec.WorkingDir
	cs.Ports = cspec.Ports
	if len(cspec.EnvFrom) != 0 {
		cs.EnvFrom = cspec.EnvFrom
	}
	if len(cspec.Env) != 0 {
		cs.Env = cspec.Env
	}
	cs.Resources = cspec.Resources
	if len(cspec.VolumeMounts) != 0 {
		cs.VolumeMounts = cspec.VolumeMounts
	}
	cs.LivenessProbe = cspec.LivenessProbe
	cs.ReadinessProbe = cspec.ReadinessProbe
	cs.Lifecycle = cspec.Lifecycle
	cs.TerminationMessagePath = cspec.TerminationMessagePath
	cs.TerminationMessagePolicy = cspec.TerminationMessagePolicy
	cs.ImagePullPolicy = cspec.ImagePullPolicy
	cs.SecurityContext = cspec.SecurityContext
	cs.StdinOnce = cspec.StdinOnce
	cs.Stdin = cspec.Stdin
	cs.TTY = cspec.TTY
	return cs

}

func V1PodToPodStatus(pod corev1.Pod) *Status {
	var s Status
	s.Name = pod.Name
	ps := pod.Status
	s.Phase = string(ps.Phase)
	s.IP = ps.PodIP
	s.Total = len(pod.Spec.Containers)
	s.HostIP = ps.HostIP
	s.Labels = make(map[string]string)
	if len(pod.Labels) != 0 {
		s.Labels = pod.Labels
	}

	s.Annotations = make(map[string]string)
	if len(pod.Annotations) != 0 {
		s.Annotations = pod.Annotations
	}

	s.RestartPolicy = string(pod.Spec.RestartPolicy)

	if ps.StartTime != nil {
		s.StartTime = ps.StartTime.Unix()
	}
	for _, v := range pod.Spec.Containers {
		cs := k8sContainerSpecTran(&v)
		s.ContainerSpecs = append(s.ContainerSpecs, *cs)

	}

	for _, v := range ps.ContainerStatuses {
		s.ContainerStatuses = append(s.ContainerStatuses, ContainerStatus{v})
		if v.Ready {
			s.Running += 1
		}

		//显示最大的重启次数
		if s.Restarts < int(v.RestartCount) {
			s.Restarts = int(v.RestartCount)
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
	/*
		if !p.memoryOnly {
			return p.Template, nil
		} else {
			return "", nil
	*/
	runtime, err := p.GetRuntimeDirect()
	if err != nil {
		return "", log.DebugPrint(err)
	}
	//pod := runtime.Pod
	//		log.DebugPrint(pod.Kind)
	//		log.DebugPrint(pod.APIVersion)
	pod := runtime.Pod
	if pod.Kind == "" {
		pod.APIVersion = "v1"
		pod.Kind = "Pod"
	}

	t, err := util.GetYamlTemplateFromObject(runtime.Pod)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	return *t, nil
}

func (p *Pod) Log(containerName string) (string, error) {
	ph, err := cluster.NewPodHandler(p.Group, p.Workspace)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	opt := cluster.LogOption{
		DisplayTailLine: 10000,
	}
	logs, err := ph.Log(p.Workspace, p.Name, containerName, opt)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	return logs, nil

}

type ContainerStat struct {
	cadvisor.ContainerStat
}

func (p *Pod) Terminal(containerName string) (string, error) {
	c, err := cluster.Controller.GetCluster(p.Group, p.Workspace)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	runtime, err := p.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	var found bool
	for _, v := range runtime.Pod.Spec.Containers {
		if v.Name == containerName {
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("container not exist in pod %v ", p.Name)
	}

	token := "1234567890987654321"

	url, err := cluster.GetTerminalUrl(p.Group, p.Workspace, p.Name, containerName, runtime.Pod.Status.HostIP, c.Name, token)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (p *Pod) Stat(containerName string) ([]ContainerStat, error) {
	runtime, err := p.GetRuntime()
	if err != nil {
		return nil, err
	}

	var containerID string

	for _, v := range runtime.Pod.Status.ContainerStatuses {
		if v.Name == containerName {
			containerID = v.ContainerID
		}
	}

	if len(containerID) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	dockerPrefix := "docker://"
	if !strings.HasPrefix(containerID, dockerPrefix) {
		return nil, fmt.Errorf("only support docker container")
	}

	id := strings.TrimPrefix(containerID, dockerPrefix)
	cadvisorID := "/docker/" + id

	cadvisorPort := "4194"
	url := "http://" + runtime.Pod.Status.HostIP + ":" + cadvisorPort + "/"

	manager, err := cadvisor.NewManager(url)
	if err != nil {
		return nil, err
	}
	cstats, err := manager.GetContainerStats(cadvisorID)
	if err != nil {
		return nil, err
	}

	stats := make([]ContainerStat, 0)
	for _, v := range cstats {
		cs := ContainerStat{v}
		stats = append(stats, cs)

	}
	return stats, nil
}

func (p *Pod) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewPodHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(p.Workspace, p.Name)
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
	//log.DebugPrint(rm)
	return rm, nil

}
