package job

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	pk "ufleet-deploy/pkg/resource/pod"

	corev1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

var (
	rm *JobManager
	/* = &JobManager{
		Groups: make(map[string]JobGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller JobController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type JobController interface {
	Create(group, workspace string, data interface{}, opt CreateOptions) error
	Delete(group, workspace, job string, opt DeleteOption) error
	Get(group, workspace, job string) (JobInterface, error)
	List(group, workspace string) ([]JobInterface, error)
	ListGroup(group string) ([]JobInterface, error)
}

type JobInterface interface {
	Info() *Job
	GetRuntime() (*Runtime, error)
	GetStatus() (*Status, error)
	GetTemplate() (string, error)
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
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	AppStack   string `json:"app"`
	User       string `json:"user"`
	Cluster    string `json:"cluster"`
	Template   string `json:"template"`
	memoryOnly bool   //用于判定pod是否由k8s自动创建
}

type GetOptions struct{}
type DeleteOption struct{}
type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/job等主动创建的资源.
	//废弃,直接通过JobManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *JobManager) get(groupName, workspaceName, jobName string) (*Job, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	job, ok := workspace.Jobs[jobName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &job, nil
}

func (p *JobManager) Get(group, workspace, jobName string) (JobInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, jobName)
}

func (p *JobManager) ListGroup(groupName string) ([]JobInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]JobInterface, 0)
	for _, v := range group.Workspaces {
		for k := range v.Jobs {
			t := v.Jobs[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *JobManager) List(groupName, workspaceName string) ([]JobInterface, error) {

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

	pis := make([]JobInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.Jobs {
		t := workspace.Jobs[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *JobManager) Create(groupName, workspaceName string, data interface{}, opt CreateOptions) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	return nil

}

//无锁
func (p *JobManager) delete(groupName, workspaceName, jobName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.Jobs, jobName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *JobManager) Delete(group, workspace, jobName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	job, err := p.get(group, workspace, jobName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if job.memoryOnly {
		ph, err := cluster.NewJobHandler(group, workspace)
		if err != nil {
			return log.DebugPrint(err)
		}

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, jobName)
		if err != nil {
			return log.DebugPrint(err)
		}
		return nil
		//TODO:ufleet创建的数据
	} else {
		return nil
	}
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
	Name        string   `json:"name"`
	User        string   `json:"user"`
	Workspace   string   `json:"workspace"`
	Group       string   `json:"group"`
	Images      []string `json:"images"`
	Containers  []string `json:"containers"`
	PodNum      int      `json:"podnum"`
	ClusterIP   string   `json:"clusterip"`
	CompleteNum int      `json:"completenum"`
	ParamNum    int      `json:"paramnum"`
	Succeeded   int      `json:"succeeded"`
	Failed      int      `json:"failed"`
	CreateTime  int64    `json:"createtime"`
	StartTime   int64    `json:"starttime"`
	Reason      string   `json:"reason"`
	//	Pods       []string `json:"pods"`
	PodStatus []pk.Status `json:"podstatus"`
}

//不包含PodStatus的信息
func K8sJobToJobStatus(job *batchv1.Job) *Status {
	var js Status
	js.Name = job.Name
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
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

	js.Succeeded = int(job.Status.Succeeded)
	js.Failed = int(job.Status.Failed)

	for _, v := range job.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
	}
	return &js

}

func (j *Job) GetStatus() (*Status, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return nil, err
	}

	js := K8sJobToJobStatus(runtime.Job)
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

func (j *Job) GetTemplate() (string, error) {
	if !j.memoryOnly {
		return j.Template, nil
	} else {
		return "", nil
	}

}

func InitJobController(be backend.BackendHandler) (JobController, error) {
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
	log.DebugPrint(rm)
	return rm, nil

}
