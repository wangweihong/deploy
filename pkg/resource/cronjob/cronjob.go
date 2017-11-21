package cronjob

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
	jk "ufleet-deploy/pkg/resource/job"
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"
	"ufleet-deploy/pkg/sign"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
)

var (
	rm         *CronJobManager
	Controller resource.ObjectController
)

type CronJobInterface interface {
	Info() *CronJob
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	GetTemplate() (string, error)
	Event() ([]corev1.Event, error)
	SuspendOrResume() error
	Metadata() resource.ObjectMeta
}

type CronJobManager struct {
	Groups map[string]CronJobGroup `json:"groups"`
	locker sync.Mutex
}

type CronJobGroup struct {
	Workspaces map[string]CronJobWorkspace `json:"Workspaces"`
}

type CronJobWorkspace struct {
	CronJobs map[string]CronJob `json:"cronjobs"`
}

type Runtime struct {
	CronJob *batchv2alpha1.CronJob
	Jobs    []*batchv1.Job
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记CronJob相关的K8s资源是否仍然存在
//在CronJob构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新CronJob的信息
type CronJob struct {
	resource.ObjectMeta
}

func (p *CronJobManager) Lock() {
	p.locker.Lock()
}
func (p *CronJobManager) Unlock() {
	p.locker.Unlock()
}
func (p *CronJobManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *CronJobManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := CronJob{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *CronJobManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*CronJob)
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
		_, ok = workspace.CronJobs[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.CronJobs[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *CronJobManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *CronJobManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group CronJobGroup
	group.Workspaces = make(map[string]CronJobWorkspace)
	p.Groups[groupName] = group
	return nil
}
func (p *CronJobManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *CronJobManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res CronJob
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *CronJobManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws CronJobWorkspace
	ws.CronJobs = make(map[string]CronJob)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewCronJobHandler(groupName, workspaceName)
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

func (p *CronJobManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

//注意这里没锁
func (p *CronJobManager) get(groupName, workspaceName, resourceName string) (*CronJob, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	cronjob, ok := workspace.CronJobs[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &cronjob, nil
}

func (p *CronJobManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *CronJobManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *CronJobManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

func (p *CronJobManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.CronJobs {
		t := workspace.CronJobs[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *CronJobManager) ListGroupObject(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.CronJobs {
			t := v.CronJobs[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *CronJobManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewCronJobHandler(groupName, workspaceName)
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
	var obj batchv2alpha1.CronJob
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must offer one  cronjob resource json/yaml data")
	}
	obj.ResourceVersion = ""
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	if opt.App != nil {
		obj.Annotations[sign.SignUfleetAppKey] = *opt.App
	}

	var cp CronJob
	cp.Name = obj.Name
	cp.Group = groupName
	cp.Kind = resourceKind
	cp.Workspace = workspaceName
	cp.Template = string(data)
	cp.User = opt.User
	cp.CreateTime = time.Now().Unix()
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

func (p *CronJobManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr batchv2alpha1.CronJob
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
	if res.App != "" {
		newr.Annotations[sign.SignUfleetAppKey] = res.App
	}

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewCronJobHandler(groupName, workspaceName)
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

//无锁
func (p *CronJobManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.CronJobs, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *CronJobManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewCronJobHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
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

func (cronjob *CronJob) Info() *CronJob {
	return cronjob
}

func (p *CronJob) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewCronJobHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, err
	}

	cj, err := ph.Get(p.Workspace, p.Name)
	if err != nil {
		return nil, err
	}
	jobs, err := ph.GetJobs(p.Workspace, p.Name)
	if err != nil {
		return nil, err
	}

	return &Runtime{CronJob: cj, Jobs: jobs}, nil
}

func (p *CronJob) GetTemplate() (string, error) {
	runtime, err := p.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.CronJob)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: batch/v2alpha1\nkind: CronJob"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)

	return *t, nil
}

func (p *CronJob) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewCronJobHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(p.Workspace, p.Name)
}

func (p *CronJob) SuspendOrResume() error {
	ph, err := cluster.NewCronJobHandler(p.Group, p.Workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	runtime, err := p.GetRuntime()
	if err != nil {
		return log.DebugPrint(err)
	}

	var suspend bool
	cronjob := runtime.CronJob

	if cronjob.Spec.Suspend == nil {
		log.DebugPrint("--------------empty")
	} else {
		log.DebugPrint(*cronjob.Spec.Suspend)

	}

	if cronjob.Spec.Suspend == nil {
		suspend = false
	} else {
		if *cronjob.Spec.Suspend {
			suspend = false
		} else {
			suspend = true
		}
	}
	log.DebugPrint(suspend)
	cronjob.ResourceVersion = ""
	cronjob.Spec.Suspend = &suspend
	err = ph.Update(p.Workspace, cronjob)
	if err != nil {
		return log.DebugPrint(err)
	}
	return nil

}

type Status struct {
	resource.ObjectMeta

	Total   int  `json:"total"`
	Active  int  `json:"active"`
	Suspend bool `json:"suspend"`
	//-1表示没有指定
	StartingDeadlineSeconds    int64              `json:"startingDeadlineSeconds"`
	SuccessfulJobsHistoryLimit int                `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     int                `json:"failedJobsHistoryLimit"`
	ConcurrencyPolicy          string             `json:"concurrencyPolicy"`
	LastScheduleTime           int64              `json:"lastscheduletime"`
	Period                     string             `json:"period"`
	Labels                     map[string]string  `json:"labels"`
	Seletors                   map[string]string  `json:"selectors"`
	JobStatus                  []jk.Status        `json:"jobstatuses"`
	ContainerSpecs             []pk.ContainerSpec `json:"containerspec"`

	Reason string `json:"reason"`
}

func (p *CronJob) ObjectStatus() resource.ObjectStatus {
	return p.GetStatus()
}

func (p *CronJob) GetStatus() *Status {
	s := Status{ObjectMeta: p.ObjectMeta}

	s.JobStatus = make([]jk.Status, 0)
	s.Labels = make(map[string]string)
	s.Seletors = make(map[string]string)
	s.ContainerSpecs = make([]pk.ContainerSpec, 0)

	runtime, err := p.GetRuntime()
	if err != nil {
		s.Reason = err.Error()
		return &s
	}

	if s.CreateTime == 0 {
		s.CreateTime = runtime.CronJob.CreationTimestamp.Unix()
	}
	info := p.Info()

	s.Period = runtime.CronJob.Spec.Schedule
	rs := runtime.CronJob.Status
	if rs.LastScheduleTime != nil {
		s.LastScheduleTime = rs.LastScheduleTime.Unix()
	}

	if runtime.CronJob.Spec.Suspend != nil {
		s.Suspend = *runtime.CronJob.Spec.Suspend
	} else {
		s.Suspend = false
	}

	s.Labels = runtime.CronJob.Spec.JobTemplate.Labels
	if runtime.CronJob.Spec.JobTemplate.Spec.Selector != nil {
		s.Seletors = runtime.CronJob.Spec.JobTemplate.Spec.Selector.MatchLabels

	}

	s.ConcurrencyPolicy = string(runtime.CronJob.Spec.ConcurrencyPolicy)
	//-1表示没有设置
	if runtime.CronJob.Spec.SuccessfulJobsHistoryLimit != nil {
		s.SuccessfulJobsHistoryLimit = int(*runtime.CronJob.Spec.SuccessfulJobsHistoryLimit)
	} else {
		s.SuccessfulJobsHistoryLimit = -1
	}

	if runtime.CronJob.Spec.FailedJobsHistoryLimit != nil {
		s.FailedJobsHistoryLimit = int(*runtime.CronJob.Spec.FailedJobsHistoryLimit)
	} else {
		s.FailedJobsHistoryLimit = -1
	}

	if runtime.CronJob.Spec.StartingDeadlineSeconds != nil {

		s.StartingDeadlineSeconds = *runtime.CronJob.Spec.StartingDeadlineSeconds
	} else {
		s.StartingDeadlineSeconds = -1

	}

	for _, v := range runtime.CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
		s.ContainerSpecs = append(s.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}

	s.Total = len(runtime.Jobs)
	s.Active = len(rs.Active)

	s.JobStatus = make([]jk.Status, 0)
	for _, v := range runtime.Jobs {
		tmp, err := jk.Controller.GetObject(info.Group, info.Workspace, v.Name)

		ji, _ := jk.GetJobInterface(tmp)

		if err != nil {
			s.Reason = err.Error()
			return &s
		}
		js := ji.GetStatus()
		s.JobStatus = append(s.JobStatus, *js)

	}
	return &s
}
func (s *CronJob) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func GetCronJobInterface(obj resource.Object) (CronJobInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*CronJob)
	if !ok {
		return nil, fmt.Errorf("resource object is not CronJob type")
	}

	return ri, nil

}
func InitCronJobController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &CronJobManager{}
	rm.Groups = make(map[string]CronJobGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group CronJobGroup
		group.Workspaces = make(map[string]CronJobWorkspace)
		for i, j := range v.Workspaces {
			var workspace CronJobWorkspace
			workspace.CronJobs = make(map[string]CronJob)
			for m, n := range j.Resources {
				var cronjob CronJob
				err := json.Unmarshal([]byte(n), &cronjob)
				if err != nil {
					return nil, fmt.Errorf("init cronjob manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.CronJobs[m] = cronjob
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
