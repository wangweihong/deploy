package cronjob

import (
	"encoding/json"
	"fmt"
	"sync"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	jk "ufleet-deploy/pkg/resource/job"
	"ufleet-deploy/pkg/resource/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
)

var (
	rm *CronJobManager
	/* = &CronJobManager{
		Groups: make(map[string]CronJobGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller CronJobController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type CronJobController interface {
	Create(group, workspace string, data []byte, opt CreateOptions) error
	Delete(group, workspace, cronjob string, opt DeleteOption) error
	Get(group, workspace, cronjob string) (CronJobInterface, error)
	List(group, workspace string) ([]CronJobInterface, error)
	ListGroup(group string) ([]CronJobInterface, error)
}

type CronJobInterface interface {
	Info() *CronJob
	GetRuntime() (*Runtime, error)
	GetStatus() (*Status, error)
	GetTemplate() (string, error)
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
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	AppStack   string `json:"app"`
	User       string `json:"user"`
	Cluster    string `json:"cluster"`
	Template   string `json:"template"`
	memoryOnly bool
}

type GetOptions struct {
}
type DeleteOption struct{}

type CreateOptions struct {
	//	MemoryOnly bool    //只在内存中创建,不创建k8s资源/也不保存在etcd中.由k8s daemonset/cronjob等主动创建的资源.
	//废弃,直接通过CronJobManager来调用
	App  *string //所属app
	User string  //创建的用户
}

//注意这里没锁
func (p *CronJobManager) get(groupName, workspaceName, cronjobName string) (*CronJob, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	cronjob, ok := workspace.CronJobs[cronjobName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &cronjob, nil
}

func (p *CronJobManager) Get(group, workspace, cronjobName string) (CronJobInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, cronjobName)
}

func (p *CronJobManager) List(groupName, workspaceName string) ([]CronJobInterface, error) {

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

	pis := make([]CronJobInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.CronJobs {
		t := workspace.CronJobs[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *CronJobManager) ListGroup(groupName string) ([]CronJobInterface, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]CronJobInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.CronJobs {
			t := v.CronJobs[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *CronJobManager) Create(groupName, workspaceName string, data []byte, opt CreateOptions) error {

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
	var cj batchv2alpha1.CronJob
	err = json.Unmarshal(exts[0].Raw, &cj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if cj.Kind != "CronJob" {
		return log.DebugPrint("must offer one  cronjob resource json/yaml data")
	}

	var cp CronJob
	cp.Name = cj.Name
	cp.Group = groupName
	cp.Workspace = workspaceName
	cp.Template = string(data)
	cp.User = opt.User
	if opt.App != nil {
		cp.AppStack = *opt.App
	}

	be := backend.NewBackendHandler()
	err = be.CreateResource(backendKind, groupName, workspaceName, cp.Name, cp)
	if err != nil {
		return log.DebugPrint(err)
	}

	err = ph.Create(workspaceName, &cj)
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
func (p *CronJobManager) delete(groupName, workspaceName, cronjobName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.CronJobs, cronjobName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *CronJobManager) Delete(group, workspace, cronjobName string, opt DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewCronJobHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	cronjob, err := p.get(group, workspace, cronjobName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if cronjob.memoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, cronjobName)
		if err != nil {
			return log.DebugPrint(err)
		}
		//TODO:ufleet创建的数据
		return nil
	} else {
		be := backend.NewBackendHandler()
		err := be.DeleteResource(backendKind, group, workspace, cronjobName)
		if err != nil {
			return log.DebugPrint(err)
		}

		err = ph.Delete(workspace, cronjobName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
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
	if !p.memoryOnly {
		return p.Template, nil
	} else {
		/*
			runtime, err := p.GetRuntime()
			if err != nil {
				return "", log.DebugPrint(err)
			}

			t, err := util.GetYamlTemplateFromObject(runtime.CronJob)
			if err != nil {
				return "", log.DebugPrint(err)
			}

			return *t, nil
		*/
		return "", nil
	}
}

type Status struct {
	Name             string      `json:"name"`
	User             string      `json:"user"`
	Workspace        string      `json:"workspace"`
	Group            string      `json:"group"`
	Total            int         `json:"total"`
	Active           int         `json:"active"`
	LastScheduleTime int64       `json:"lastscheduletime"`
	Period           string      `json:"period"`
	JobStatus        []jk.Status `json:"jobstatuses"`
	Reason           string      `json:"reason"`
}

/*
func K8sCronJobToStatus(job *batchv2alpha1.CronJob) *Status {
	var js Status
	js.Name = job.Name
	js.Period = job.Spec.Schedule
	if job.Status.LastScheduleTime != nil {
		js.LastScheduleTime = job.Status.LastScheduleTime
	}

}
*/

func (p *CronJob) GetStatus() (*Status, error) {
	var s Status
	runtime, err := p.GetRuntime()
	if err != nil {
		return nil, err
	}
	info := p.Info()

	s.Name = info.Name
	s.User = info.User
	s.Group = info.Group
	s.Workspace = info.Workspace

	s.Period = runtime.CronJob.Spec.Schedule
	rs := runtime.CronJob.Status
	if rs.LastScheduleTime != nil {
		s.LastScheduleTime = rs.LastScheduleTime.Unix()
	}

	s.Total = len(runtime.Jobs)
	s.Active = len(rs.Active)

	s.JobStatus = make([]jk.Status, 0)
	for _, v := range runtime.Jobs {
		//		js := jk.K8sJobToJobStatus(v)
		//		s.JobStatus = append(s.JobStatus, *js)

		ji, err := jk.Controller.Get(info.Group, info.Workspace, v.Name)
		if err != nil {
			s.Reason = err.Error()
			return nil, err
		}
		js, err := ji.GetStatus()
		if err != nil {
			return nil, err
		}
		s.JobStatus = append(s.JobStatus, *js)

	}
	return &s, nil
}

func InitCronJobController(be backend.BackendHandler) (CronJobController, error) {
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
	log.DebugPrint(rm)
	return rm, nil

}
