package deployment

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

	pk "ufleet-deploy/pkg/resource/pod"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/pkg/api"
	corev1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm         *DeploymentManager
	Controller resource.ObjectController
)

type DeploymentInterface interface {
	Info() *Deployment
	GetRuntime() (*Runtime, error)
	GetRuntimeObjectCopy() (*extensionsv1beta1.Deployment, error)
	GetStatus() *Status
	Scale(num int) error
	Event() ([]corev1.Event, error)
	GetTemplate() (string, error)
	GetAllReplicaSets() (map[int64]*extensionsv1beta1.ReplicaSet, error)
	GetRevisionsAndDescribe() (map[int64]string, error)
	Rollback(revision int64) (*string, error)
	GetAutoScale() (*HPA, error)
	StartAutoScale(min int, max int, cpuPercent int, memPercent int, diskPercent int, NetPercent int) error
	ResumeOrPauseRollOut() error
	GetServices() ([]*corev1.Service, error)
}

type DeploymentManager struct {
	Groups map[string]DeploymentGroup `json:"groups"`
	locker sync.Mutex
}

type DeploymentGroup struct {
	Workspaces map[string]DeploymentWorkspace `json:"Workspaces"`
}

type DeploymentWorkspace struct {
	Deployments map[string]Deployment `json:"deployments"`
}

type Runtime struct {
	Deployment *extensionsv1beta1.Deployment
	Pods       []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Deployment相关的K8s资源是否仍然存在
//在Deployment构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Deployment的信息
type Deployment struct {
	resource.ObjectMeta
	AutoScaler HPA `json:"autoscaler"`
}

type HPA struct {
	Deployed      bool `json:"deployed"`
	CpuPercernt   int  `json:"cpuPercent"`
	MemoryPercent int  `json:"memPercent"`
	DiskPercent   int  `json:"diskPercent"`
	NetPercent    int  `json:"netPercent"`
	MinReplicas   int  `json:"minReplicas"`
	MaxReplicas   int  `json:"maxReplicas"`
}

func GetDeploymentInterface(obj resource.Object) (DeploymentInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Deployment)
	if !ok {
		return nil, fmt.Errorf("resource object is not deployment type")
	}

	return ri, nil
}

func (p *DeploymentManager) Lock() {
	p.locker.Lock()
}

func (p *DeploymentManager) Unlock() {
	p.locker.Unlock()
}

func (p *DeploymentManager) Kind() string {
	return resourceKind

}

//仅仅用于基于内存的对象的创建
func (p *DeploymentManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Deployment{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *DeploymentManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Deployment)
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
		_, ok = workspace.Deployments[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Deployments[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *DeploymentManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *DeploymentManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group DeploymentGroup
	group.Workspaces = make(map[string]DeploymentWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *DeploymentManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *DeploymentManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Deployment
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *DeploymentManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws DeploymentWorkspace
	ws.Deployments = make(map[string]Deployment)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewDeploymentHandler(groupName, workspaceName)
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

func (p *DeploymentManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *DeploymentManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
	//注意这里没锁
}

func (p *DeploymentManager) get(groupName, workspaceName, resourceName string) (*Deployment, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	deployment, ok := workspace.Deployments[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &deployment, nil
}

func (p *DeploymentManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *DeploymentManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *DeploymentManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.Deployments {
		t := workspace.Deployments[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *DeploymentManager) ListGroupObject(groupName string) ([]resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)
	for _, v := range group.Workspaces {
		for k := range v.Deployments {
			t := v.Deployments[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *DeploymentManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewDeploymentHandler(groupName, workspaceName)
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

	var obj extensionsv1beta1.Deployment
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}

	obj.ResourceVersion = ""

	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	obj.Annotations[sign.SignUfleetAutoScaleSupported] = "true"

	if obj.Spec.Template.Annotations == nil {
		obj.Spec.Template.Annotations = make(map[string]string)
	}
	if opt.App != nil {
		obj.Spec.Template.Annotations[sign.SignUfleetAppKey] = *opt.App
	}
	obj.Spec.Template.Annotations[sign.SignUfleetAutoScaleSupported] = "true"
	obj.Spec.Template.Annotations[sign.SignUfleetDeployment] = obj.Name

	var cp Deployment
	cp.CreateTime = time.Now().Unix()
	cp.Kind = resourceKind
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	cp.App = resource.DefaultAppBelong
	if opt.App != nil {
		cp.App = *opt.App
	}
	cp.User = opt.User
	//因为pod创建时,触发informer,所以优先创建etcd
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
func (p *DeploymentManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Deployments, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DeploymentManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewDeploymentHandler(group, workspace)
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
func (p *DeploymentManager) update(groupName, workspaceName string, resourceName string, d *Deployment) error {

	//	p.locker.Lock()
	//	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	_, ok = workspace.Deployments[resourceName]
	if !ok {
		return resource.ErrResourceNotFound
	}
	workspace.Deployments[resourceName] = *d
	group.Workspaces[workspaceName] = workspace
	rm.Groups[groupName] = group
	return nil
}

func (p *DeploymentManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.Deployment
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
		newr.Annotations[sign.SignUfleetAutoScaleSupported] = "true"
	}

	if newr.Spec.Template.Annotations == nil {
		newr.Spec.Template.Annotations = make(map[string]string)
	}
	if res.App != "" {
		newr.Spec.Template.Annotations[sign.SignUfleetAppKey] = res.App
	}
	newr.Spec.Template.Annotations[sign.SignUfleetDeployment] = newr.Name
	newr.Spec.Template.Annotations[sign.SignUfleetAutoScaleSupported] = "true"

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewDeploymentHandler(groupName, workspaceName)
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

func (deployment *Deployment) Info() *Deployment {
	return deployment
}

func (j *Deployment) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	deployment, err := ph.Get(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	pods, err := ph.GetPods(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	var runtime Runtime
	runtime.Pods = pods
	runtime.Deployment = deployment
	return &runtime, nil
}

type Status struct {
	resource.ObjectMeta

	Images      []string           `json:"images"`
	Containers  []string           `json:"containers"`
	PodNum      int                `json:"podnum"`
	ClusterIP   string             `json:"clusterip"`
	Strategy    string             `json:"strategy"`
	Desire      int                `json:"desire"`
	Current     int                `json:"current"`
	Available   int                `json:"available"`
	UpToDate    int                `json:"uptodate"`
	Ready       int                `json:"ready"`
	Labels      map[string]string  `json:"labels"`
	Annotations map[string]string  `json:"annotations"`
	Selectors   map[string]string  `json:"selectors"`
	Reason      string             `json:"reason"`
	timeout     int64              `json:"timemout"`
	Paused      bool               `json:"paused"`
	Revision    int64              `json:"revision"`
	ReplicaSet  string             `json:"replicaset"`
	PodsCount   resource.PodsCount `json:"podscount"`
	//	Pods       []string `json:"pods"`
	PodStatus               []pk.Status        `json:"podstatus"`
	ContainerSpecs          []pk.ContainerSpec `json:"containerspec"`
	ProgressDeadlineSeconds int                `json:"progressdeadlineseconds"`
	extensionsv1beta1.DeploymentStatus
}

func (j *Deployment) ObjectStatus() resource.ObjectStatus {
	return j.GetStatus()
}

func (j *Deployment) GetStatus() *Status {
	var e error
	var js *Status
	var deployment *extensionsv1beta1.Deployment
	var rev *int64
	var rs *extensionsv1beta1.ReplicaSet
	var ph cluster.DeploymentHandler
	runtime, err := j.GetRuntime()
	if err != nil {
		e = err
		goto faileReturn
	}

	ph, err = cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		e = err
		goto faileReturn
	}

	log.DebugPrint("get current revesion and rs")
	rev, rs, err = ph.GetCurrentRevisionAndReplicaSet(j.Workspace, j.Name)
	if err != nil {
		e = err
		goto faileReturn
	}

	deployment = runtime.Deployment
	js = &Status{ObjectMeta: j.ObjectMeta, DeploymentStatus: deployment.Status}
	//js = &Status{ObjectMeta: j.ObjectMeta} //, DeploymentStatus: deployment.Status}
	if rev != nil {
		js.Revision = *rev
	} else {
		js.Revision = 0
	}
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Labels = make(map[string]string)
	js.Annotations = make(map[string]string)
	js.Selectors = make(map[string]string)
	js.Paused = deployment.Spec.Paused
	js.Strategy = string(deployment.Spec.Strategy.Type)
	if rs != nil {
		js.ReplicaSet = rs.Name
	}

	if js.CreateTime == 0 {
		js.CreateTime = deployment.CreationTimestamp.Unix()
	}

	if deployment.Labels != nil {
		js.Labels = deployment.Labels
	}
	if deployment.Annotations != nil {
		js.Annotations = deployment.Annotations
	}

	if deployment.Spec.ProgressDeadlineSeconds != nil {
		js.ProgressDeadlineSeconds = int(*deployment.Spec.ProgressDeadlineSeconds)
	} else {
		js.ProgressDeadlineSeconds = -1
	}

	if deployment.Spec.Selector != nil {
		js.Selectors = deployment.Spec.Selector.MatchLabels
	}

	if deployment.Spec.Replicas != nil {
		js.Desire = int(*deployment.Spec.Replicas)
	} else {
		js.Desire = 1

	}
	js.Current = int(deployment.Status.Replicas)
	js.Ready = int(deployment.Status.ReadyReplicas)
	js.Available = int(deployment.Status.AvailableReplicas)
	js.UpToDate = int(deployment.Status.UpdatedReplicas)

	js.PodsCount = *resource.GetPodsCount(runtime.Pods)

	for _, v := range deployment.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}
	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
	}

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)
	}

	return js
faileReturn:
	js = &Status{ObjectMeta: j.ObjectMeta}
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Labels = make(map[string]string)
	js.Annotations = make(map[string]string)
	js.Selectors = make(map[string]string)
	js.Reason = e.Error()
	return js
}

func (j *Deployment) Scale(num int) error {

	jh, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return err
	}

	err = jh.Scale(j.Workspace, j.Name, int32(num))
	if err != nil {
		return err
	}

	return nil
}
func (j *Deployment) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(j.Workspace, j.Name)
}

func (j *Deployment) GetTemplate() (string, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.Deployment)
	if err != nil {
		return "", log.DebugPrint(err)
	}
	prefix := "apiVersion: extensions/v1beta1\nkind: Deployment"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)

	return *t, nil
}

func (j *Deployment) GetAllReplicaSets() (map[int64]*extensionsv1beta1.ReplicaSet, error) {
	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	rm, err := ph.GetRevisionsAndReplicas(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	return rm, nil

}

func (j *Deployment) GetRevisionsAndDescribe() (map[int64]string, error) {
	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	rm, err := ph.GetRevisionsAndDescribe(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	rs := make(map[int64]string, 0)
	for k, v := range rm {
		str, err := json.Marshal(v)
		if err != nil {
			return nil, log.DebugPrint(err)
		}
		rs[k] = string(str)
	}

	return rs, nil
}

func (j *Deployment) Rollback(revision int64) (*string, error) {
	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	return ph.Rollback(j.Workspace, j.Name, revision)

}

//需要加锁
func (j *Deployment) StartAutoScale(min int, max int, cpuPercent int, memPercent int, diskPercent int, NetPercent int) error {

	j.AutoScaler.Deployed = true
	j.AutoScaler.CpuPercernt = cpuPercent
	j.AutoScaler.MemoryPercent = memPercent
	j.AutoScaler.NetPercent = NetPercent
	j.AutoScaler.DiskPercent = diskPercent
	j.AutoScaler.MaxReplicas = max
	j.AutoScaler.MinReplicas = min

	if j.MemoryOnly != true {

		be := backend.NewBackendHandler()
		err := be.UpdateResource(backendKind, j.Group, j.Workspace, j.Name, j)
		if err != nil {
			return log.DebugPrint(err)
		}
	} else {
		return fmt.Errorf("deployment is not created by ufleet directly doesn't support autoscale")
	}
	return nil
}

func (j *Deployment) GetAutoScale() (*HPA, error) {
	return &j.AutoScaler, nil
}

func (j *Deployment) ResumeOrPauseRollOut() error {
	runtime, err := j.GetRuntime()
	if err != nil {
		return err
	}

	ph, err := cluster.NewDeploymentHandler(j.Group, j.Workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	if runtime.Deployment.Spec.Paused {
		return ph.ResumeRollout(j.Workspace, j.Name)
	} else {
		return ph.PauseRollout(j.Workspace, j.Name)
	}
}

func (p *Deployment) GetServices() ([]*corev1.Service, error) {
	ph, err := cluster.NewDeploymentHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.GetServices(p.Workspace, p.Name)
}

func (s *Deployment) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func (p *Deployment) GetRuntimeObjectCopy() (*extensionsv1beta1.Deployment, error) {
	r, err := p.GetRuntime()
	if err != nil {
		return nil, err
	}

	newobj, err := api.Scheme.Copy(r.Deployment)
	if err != nil {
		return nil, err
	}

	newd, ok := newobj.(*extensionsv1beta1.Deployment)
	if !ok {
		err := fmt.Errorf("deep copy object fail, object type doesn't match")
		return nil, err
	}

	return newd, nil
}

func InitDeploymentController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &DeploymentManager{}
	rm.Groups = make(map[string]DeploymentGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group DeploymentGroup
		group.Workspaces = make(map[string]DeploymentWorkspace)
		for i, j := range v.Workspaces {
			var workspace DeploymentWorkspace
			workspace.Deployments = make(map[string]Deployment)
			for m, n := range j.Resources {
				var deployment Deployment
				err := json.Unmarshal([]byte(n), &deployment)
				if err != nil {
					return nil, fmt.Errorf("init deployment manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Deployments[m] = deployment
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
