package job

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
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"
	"ufleet-deploy/pkg/sign"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

var (
	rm *JobManager

	Controller resource.ObjectController
)

type JobInterface interface {
	Info() *Job
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	GetTemplate() (string, error)
	Event() ([]corev1.Event, error)
	//	Runtime()
}

type JobManager struct {
	Groups map[string]JobGroup `json:"groups"`
	locker sync.Mutex
}

type JobGroup struct {
	Workspaces map[string]JobWorkspace `json:"Workspaces"`
}

type JobWorkspace struct {
	Jobs map[string]Job `json:"jobs"`
}

type Runtime struct {
	Job  *batchv1.Job
	Pods []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Job相关的K8s资源是否仍然存在
//在Job构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Job的信息
type Job struct {
	resource.ObjectMeta
}

func GetJobInterface(obj resource.Object) (JobInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Job)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *JobManager) Lock() {
	p.locker.Lock()
}

func (p *JobManager) Unlock() {
	p.locker.Unlock()
}

func (p *JobManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *JobManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Job{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *JobManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Job)
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
		_, ok = workspace.Jobs[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Jobs[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *JobManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *JobManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group JobGroup
	group.Workspaces = make(map[string]JobWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *JobManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *JobManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Job
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *JobManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws JobWorkspace
	ws.Jobs = make(map[string]Job)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewJobHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	res, err := ph.List(workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	for _, e := range res {

		var o resource.ObjectMeta
		o.Name = e.Name
		o.MemoryOnly = true
		o.Workspace = workspaceName
		o.Group = groupName
		o.User = "kubernetes"

		err = p.NewObject(o)
		if err != nil && err != resource.ErrResourceExists {
			return log.ErrorPrint(err)
		}
	}
	return nil

}

func (p *JobManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *JobManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *JobManager) get(groupName, workspaceName, resourceName string) (*Job, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	job, ok := workspace.Jobs[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &job, nil
}

func (p *JobManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *JobManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *JobManager) ListGroupObject(groupName string) ([]resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)
	for _, v := range group.Workspaces {
		for k := range v.Jobs {
			t := v.Jobs[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *JobManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.Jobs {
		t := workspace.Jobs[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *JobManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewJobHandler(groupName, workspaceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	exts, err := util.ParseJsonOrYaml(data)
	if err != nil {
		return log.DebugPrint(err)
	}
	if len(exts) != 1 {
		return log.DebugPrint("must offer one  resource json/yaml data")
	}
	var obj batchv1.Job
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must offer one  resource json/yaml data")
	}
	obj.ResourceVersion = ""
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp Job
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Group = groupName
	cp.Workspace = workspaceName
	cp.Template = string(data)
	cp.User = opt.User
	cp.Kind = resourceKind
	cp.App = resource.DefaultAppBelong
	if opt.App != nil {
		cp.App = *opt.App
	}

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

//无锁
func (p *JobManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Jobs, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *JobManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewJobHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	res, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.MemoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, resourceName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
		}
		return nil
		//TODO:ufleet创建的数据
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
		if !opt.DontCallApp && res.App != resource.DefaultAppBelong {
			go func() {
				var re resource.ResourceEvent
				re.Group = group
				re.Workspace = workspace
				re.Kind = resourceKind
				re.Action = resource.ResourceActionDelete
				re.Resource = res.Name
				re.App = res.App

				resource.ResourceEventChan <- re
			}()
		}
		return nil

	}
}

func (p *JobManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr batchv1.Job
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""
	if !res.MemoryOnly {
		if newr.Annotations == nil {
			newr.Annotations = make(map[string]string)
		}
		newr.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	}

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewJobHandler(groupName, workspaceName)
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
		return err
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

func (j *Job) Info() *Job {
	return j
}

func (j *Job) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewJobHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	job, err := ph.Get(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	pods, err := ph.GetPods(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	var runtime Runtime
	runtime.Pods = pods
	runtime.Job = job
	return &runtime, nil
}

type Status struct {
	resource.ObjectMeta
	Images      []string          `json:"images"`
	Containers  []string          `json:"containers"`
	PodNum      int               `json:"podnum"`
	ClusterIP   string            `json:"clusterip"`
	CompleteNum int               `json:"completenum"`
	ParamNum    int               `json:"paramnum"`
	Succeeded   int               `json:"succeeded"`
	Failed      int               `json:"failed"`
	CreateTime  int64             `json:"createtime"`
	StartTime   int64             `json:"starttime"`
	Labels      map[string]string `json:"labels"`
	Selector    map[string]string `json:"selector"`
	Reason      string            `json:"reason"`
	//-1表示没有设置
	ActiveDeadlineSeconds int64 `json:"activedeadlineseconds"`
	//	Pods       []string `json:"pods"`
	PodStatus        []pk.Status        `json:"podstatus"`
	ContainerSpecs   []pk.ContainerSpec `json:"containerspec"`
	CreatorReference *CreatorReference  `json:"creatorreference"`
}

type CreatorReference struct {
	corev1.SerializedReference
	Workspace string `json:"workspace"`
	Group     string `json:"group"`
}

//不包含PodStatus的信息
func K8sJobToJobStatus(job *batchv1.Job) *Status {
	var js Status
	js.Name = job.Name
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.Labels = make(map[string]string)
	js.Selector = make(map[string]string)

	js.Labels = job.Labels
	if job.Spec.Selector != nil {
		js.Selector = job.Spec.Selector.MatchLabels
	}
	js.CreateTime = job.CreationTimestamp.Unix()
	if job.Spec.Parallelism != nil {
		js.ParamNum = int(*job.Spec.Parallelism)
	}
	if job.Spec.Completions != nil {
		js.CompleteNum = int(*job.Spec.Completions)
	}

	if job.Status.StartTime != nil {
		js.StartTime = job.Status.StartTime.Unix()
	}

	if job.Spec.ActiveDeadlineSeconds != nil {
		js.ActiveDeadlineSeconds = *job.Spec.ActiveDeadlineSeconds
	} else {
		//表示没有设置
		js.ActiveDeadlineSeconds = -1
	}

	js.Succeeded = int(job.Status.Succeeded)
	js.Failed = int(job.Status.Failed)

	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	for _, v := range job.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}
	return &js

}
func (j *Job) ObjectStatus() resource.ObjectStatus {

	return j.GetStatus()
}

func (j *Job) GetStatus() *Status {
	var err error
	var e error
	var ph cluster.JobHandler
	var sr *corev1.SerializedReference
	var js *Status

	runtime, e := j.GetRuntime()
	if err != nil {
		err = e
		goto MeetError
	}

	ph, err = cluster.NewJobHandler(j.Group, j.Workspace)
	if err != nil {
		goto MeetError
	}

	sr, err = ph.GetCreator(j.Workspace, j.Name)
	if err != nil {
		goto MeetError
	}

	js = K8sJobToJobStatus(runtime.Job)

	if sr != nil {
		js.CreatorReference = &CreatorReference{SerializedReference: *sr, Group: j.Group, Workspace: j.Workspace}
	}
	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
	}

	js.ObjectMeta = j.ObjectMeta

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)
	}

	return js
MeetError:
	js = &Status{ObjectMeta: j.ObjectMeta}
	js.Name = j.Name
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.Labels = make(map[string]string)
	js.Selector = make(map[string]string)

	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Reason = err.Error()
	return js
}

func (j *Job) GetTemplate() (string, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.Job)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	prefix := "apiVersion: batch/v1\nkind: Job"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)

	return *t, nil

}
func (j *Job) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewJobHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(j.Workspace, j.Name)
}

func (s *Job) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}
func InitJobController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &JobManager{}
	rm.Groups = make(map[string]JobGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group JobGroup
		group.Workspaces = make(map[string]JobWorkspace)
		for i, j := range v.Workspaces {
			var workspace JobWorkspace
			workspace.Jobs = make(map[string]Job)
			for m, n := range j.Resources {
				var job Job
				err := json.Unmarshal([]byte(n), &job)
				if err != nil {
					return nil, fmt.Errorf("init job manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Jobs[m] = job
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
