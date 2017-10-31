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
	"ufleet-deploy/pkg/resource"
	"ufleet-deploy/pkg/resource/util"
	"ufleet-deploy/pkg/sign"
	cadvisor "ufleet-deploy/util/cadvisor"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apis/pkg/api/errors"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm         *PodManager
	Controller resource.ObjectController
)

type PodInterface interface {
	Info() *Pod
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
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
	resource.ObjectMeta
}

func GetPodInterface(obj resource.Object) (PodInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Pod)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *PodManager) Lock() {
	p.locker.Lock()
}
func (p *PodManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *PodManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Pod{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *PodManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Pod)
	if !ok {
		return fmt.Errorf("object is not correct type")
	}

	group, ok := rm.Groups[cm.Group]
	if !ok {
		return resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[cm.Workspace]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	if !force {
		_, ok = workspace.Pods[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Pods[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *PodManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *PodManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group PodGroup
	group.Workspaces = make(map[string]PodWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *PodManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Pod
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *PodManager) AddWorkspace(groupName string, workspaceName string) error {
	p.Lock()
	defer p.Unlock()
	g, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	_, ok = g.Workspaces[workspaceName]
	if ok {
		return resource.ErrWorkspaceExists
	}

	var ws PodWorkspace
	ws.Pods = make(map[string]Pod)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *PodManager) DeleteWorkspace(groupName string, workspaceName string) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	_, ok = group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}
	delete(group.Workspaces, workspaceName)
	p.Groups[groupName] = group
	return nil
}

func (p *PodManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *PodManager) get(groupName, workspaceName, podName string) (*Pod, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	pod, ok := workspace.Pods[podName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &pod, nil
}

func (p *PodManager) GetObject(group, workspace, podName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, podName)
}

func (p *PodManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, fmt.Errorf("%v:group/%v,workspace/%v", resource.ErrWorkspaceNotFound, groupName, workspaceName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Pods {
		t := workspace.Pods[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *PodManager) ListGroup(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Pods {
			t := v.Pods[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *PodManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var obj corev1.Pod
	obj.Annotations = make(map[string]string)
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	obj.ResourceVersion = ""
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp Pod
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	cp.Comment = opt.Comment
	cp.Kind = resourceKind
	cp.App = resource.DefaultAppBelong
	if opt.App != nil {
		cp.App = *opt.App
	}
	cp.User = opt.User
	//因为obj创建时,触发informer,所以优先创建etcd
	be := backend.NewBackendHandler()
	err = be.CreateResource(backendKind, groupName, workspaceName, cp.Name, cp)
	if err != nil {
		return log.DebugPrint(err)
	}

	err = ph.Create(workspaceName, &obj)
	if err != nil {
		err2 := be.DeleteResource(backendKind, groupName, workspaceName, cp.Name)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil

}

func (p *PodManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	exts, err := util.ParseJsonOrYaml(data)
	if err != nil {
		return log.DebugPrint(err)
	}
	if len(exts) != 1 {
		return log.DebugPrint("must  offer only one  resource json/yaml data")
	}

	//说明是主动创建的..
	var newr corev1.Pod
	//	err = util.GetObjectFromYamlTemplate(data, &newr)
	//	err = util.GetObjectFromYamlTemplate(exts[0].Raw, &newr)
	err = json.Unmarshal(exts[0].Raw, &newr)
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

	if res.MemoryOnly {
		err = ph.Update(workspaceName, &newr)
		if err != nil {
			return log.DebugPrint(err)
		}
		return nil
	}

	old := *res
	res.Comment = opt.Comment
	be := backend.NewBackendHandler()
	err = be.UpdateResource(backendKind, res.Group, res.Workspace, res.Name, res)
	if err != nil {
		return log.DebugPrint(err)
	}

	err = ph.Update(workspaceName, &newr)
	if err != nil {
		err2 := be.UpdateResource(backendKind, res.Group, res.Workspace, res.Name, &old)
		if err2 != nil {
			log.ErrorPrint(err2)
		}
		return log.DebugPrint(err)
	}

	return nil
}

//无锁
func (p *PodManager) delete(groupName, workspaceName, podName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Pods, podName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *PodManager) DeleteObject(group, workspace, podName string, opt resource.DeleteOption) error {
	ph, err := cluster.NewPodHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, podName)
	}

	pod, err := p.get(group, workspace, podName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if pod.MemoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, podName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
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
		if !opt.DontCallApp && pod.App != resource.DefaultAppBelong {
			go func() {
				var re resource.ResourceEvent
				re.Group = group
				re.Workspace = workspace
				re.Kind = resourceKind
				re.Action = resource.ResourceActionDelete
				re.Resource = podName
				re.App = pod.App

				resource.ResourceEventChan <- re
			}()
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
	CreateTime        int64             `json:"createtime"`
	Running           int               `json:"running"`
	Total             int               `json:"total"`
	ID                string            `json:"id"`
	Reason            string            `json:"reason"`
	Restarts          int               `json:"restarts"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	RestartPolicy     string            `json:"restartpolicy"`
	ContainerStatuses []ContainerStatus `json:"containerstatuses"`
	ContainerSpecs    []ContainerSpec   `json:"containerspec"`
	Containers        []string          `json:"containers"`
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
	LivenessProbe            *Probe                          `json:"livenessProbe"`
	ReadinessProbe           *Probe                          `json:"readinessProbe"`
	Lifecycle                *corev1.Lifecycle               `json:"lifecycle"`
	TerminationMessagePath   string                          `json:"terminationMessagePath"`
	TerminationMessagePolicy corev1.TerminationMessagePolicy `json:"terminationMessagePolicy"`
	ImagePullPolicy          corev1.PullPolicy               `json:"imagePullPolicy"`
	SecurityContext          *corev1.SecurityContext         `json:"securityContext"`
	StdinOnce                bool                            `json:"stdinOnce"`
	Stdin                    bool                            `json:"stdin"`
	TTY                      bool                            `json:"tty"`
}
type Probe struct {
	*corev1.Probe
	Type string `json:"type"`
}

func K8sContainerSpecTran(cspec *corev1.Container) *ContainerSpec {
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
	//cs.LivenessProbe = cspec.LivenessProbe
	//cs.ReadinessProbe = cspec.ReadinessProbe
	var lp Probe
	if cspec.LivenessProbe != nil {
		if cspec.LivenessProbe.Exec != nil {
			lp = Probe{Probe: cspec.LivenessProbe, Type: "EXEC"}
		}
		if cspec.LivenessProbe.HTTPGet != nil {
			lp = Probe{Probe: cspec.LivenessProbe, Type: "HTTP"}
		}
		if cspec.LivenessProbe.TCPSocket != nil {
			lp = Probe{Probe: cspec.LivenessProbe, Type: "TCP"}
		}

		cs.LivenessProbe = &lp
	} else {
		cs.LivenessProbe = nil
	}

	var rp Probe
	if cspec.ReadinessProbe != nil {
		if cspec.ReadinessProbe.Exec != nil {
			rp = Probe{Probe: cspec.ReadinessProbe, Type: "EXEC"}
		}
		if cspec.ReadinessProbe.HTTPGet != nil {
			rp = Probe{Probe: cspec.ReadinessProbe, Type: "HTTP"}
		}
		if cspec.ReadinessProbe.TCPSocket != nil {
			rp = Probe{Probe: cspec.ReadinessProbe, Type: "TCP"}
		}

		cs.LivenessProbe = &rp
	} else {
		cs.LivenessProbe = nil
	}

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
	s.Containers = make([]string, 0)
	s.ID = string(pod.UID)
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
		s.CreateTime = ps.StartTime.Unix()
	}
	for _, v := range pod.Spec.Containers {
		cs := K8sContainerSpecTran(&v)
		s.ContainerSpecs = append(s.ContainerSpecs, *cs)
		s.Containers = append(s.Containers, v.Name)

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

func (p *Pod) ObjectStatus() resource.ObjectStatus {
	return p.GetStatus()
}

func (p *Pod) GetStatus() *Status {
	runtime, err := p.GetRuntime()
	if err != nil {
		var s Status
		s.Name = p.Name
		s.Containers = make([]string, 0)
		s.Labels = make(map[string]string)
		s.Reason = err.Error()
		return &s
	}
	s := V1PodToPodStatus(*runtime.Pod)

	return s
}

func (p *Pod) GetTemplate() (string, error) {
	runtime, err := p.GetRuntimeDirect()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.Pod)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	prefix := "apiVersion: v1\nkind: Pod"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)

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
func (s *Pod) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitPodController(be backend.BackendHandler) (resource.ObjectController, error) {
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
