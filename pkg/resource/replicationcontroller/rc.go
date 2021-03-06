package replicationcontroller

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

	corev1 "k8s.io/api/core/v1"
)

var (
	rm         *ReplicationControllerManager
	Controller resource.ObjectController
)

type ReplicationControllerInterface interface {
	Info() *ReplicationController
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	Scale(num int) error
	Event() ([]corev1.Event, error)
	GetTemplate() (string, error)
	GetServices() ([]*corev1.Service, error)
	GetRuntimeObjectCopy() (*corev1.ReplicationController, error)
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
	resource.ObjectMeta
}

func GetReplicationControllerInterface(obj resource.Object) (ReplicationControllerInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*ReplicationController)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *ReplicationControllerManager) Lock() {
	p.locker.Lock()
}

func (p *ReplicationControllerManager) Unlock() {
	p.locker.Unlock()
}

func (p *ReplicationControllerManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *ReplicationControllerManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := ReplicationController{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *ReplicationControllerManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*ReplicationController)
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
		_, ok = workspace.ReplicationControllers[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.ReplicationControllers[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *ReplicationControllerManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *ReplicationControllerManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group ReplicationControllerGroup
	group.Workspaces = make(map[string]ReplicationControllerWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *ReplicationControllerManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return gs
}

func (p *ReplicationControllerManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res ReplicationController
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *ReplicationControllerManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws ReplicationControllerWorkspace
	ws.ReplicationControllers = make(map[string]ReplicationController)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewReplicationControllerHandler(groupName, workspaceName)
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
		o.Kind = resourceKind

		err = p.NewObject(o)
		if err != nil && err != resource.ErrResourceExists {
			return log.ErrorPrint(err)
		}
	}
	return nil

}

func (p *ReplicationControllerManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *ReplicationControllerManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *ReplicationControllerManager) get(groupName, workspaceName, replicationcontrollerName string) (*ReplicationController, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	replicationcontroller, ok := workspace.ReplicationControllers[replicationcontrollerName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &replicationcontroller, nil
}

func (p *ReplicationControllerManager) GetObject(group, workspace, replicationcontrollerName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, replicationcontrollerName)
}

func (p *ReplicationControllerManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *ReplicationControllerManager) ListGroupObject(groupName string) ([]resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)
	for _, v := range group.Workspaces {
		for k := range v.ReplicationControllers {
			t := v.ReplicationControllers[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *ReplicationControllerManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.ReplicationControllers {
		t := workspace.ReplicationControllers[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ReplicationControllerManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var obj corev1.ReplicationController
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one rc json/yaml data")
	}
	obj.ResourceVersion = ""
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	if opt.App != nil {
		obj.Annotations[sign.SignUfleetAppKey] = *opt.App
		if obj.Spec.Template.Annotations == nil {
			obj.Spec.Template.Annotations = make(map[string]string)
		}
		obj.Spec.Template.Annotations[sign.SignUfleetAppKey] = *opt.App
	}

	var cp ReplicationController
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	cp.App = resource.DefaultAppBelong
	cp.Comment = opt.Comment
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
func (p *ReplicationControllerManager) delete(groupName, workspaceName, replicationcontrollerName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.ReplicationControllers, replicationcontrollerName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ReplicationControllerManager) DeleteObject(group, workspace, replicationcontrollerName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, replicationcontrollerName)
	}

	ph, err := cluster.NewReplicationControllerHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	res, err := p.get(group, workspace, replicationcontrollerName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.MemoryOnly {

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

func (p *ReplicationControllerManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.ReplicationController
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
		if newr.Spec.Template.Annotations == nil {
			newr.Spec.Template.Annotations = make(map[string]string)
		}
		newr.Spec.Template.Annotations[sign.SignUfleetAppKey] = res.App
	}

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewReplicationControllerHandler(groupName, workspaceName)
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
	resource.ObjectMeta
	Images     []string `json:"images"`
	Containers []string `json:"containers"`
	PodNum     int      `json:"podnum"`
	ClusterIP  string   `json:"clusterip"`
	CreateTime int64    `json:"createtime"`
	//Replicas    int32             `json:"replicas"`
	Desire      int               `json:"desire"`
	Current     int               `json:"current"`
	Available   int               `json:"available"`
	Ready       int               `json:"ready"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Selectors   map[string]string `json:"selectors"`
	Reason      string            `json:"reason"`
	//	Pods       []string `json:"pods"`
	ContainerSpecs []pk.ContainerSpec `json:"containerspec"`
	PodStatus      []pk.Status        `json:"podstatus"`
	PodsCount      resource.PodsCount `json:"podscount"`
	corev1.ReplicationControllerStatus
}

//不包含PodStatus的信息
func K8sReplicationControllerToReplicationControllerStatus(replicationcontroller *corev1.ReplicationController) *Status {
	js := Status{ReplicationControllerStatus: replicationcontroller.Status}
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
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

	js.Annotations = make(map[string]string)
	if replicationcontroller.Annotations != nil {
		js.Annotations = replicationcontroller.Annotations
	}

	js.Selectors = make(map[string]string)
	if replicationcontroller.Spec.Selector != nil {
		js.Selectors = replicationcontroller.Spec.Selector
	}

	if replicationcontroller.Spec.Replicas != nil {
		js.Desire = int(*replicationcontroller.Spec.Replicas)
	} else {
		js.Desire = 1

	}
	js.Current = int(replicationcontroller.Status.AvailableReplicas)
	js.Ready = int(replicationcontroller.Status.ReadyReplicas)
	js.Available = int(replicationcontroller.Status.AvailableReplicas)

	for _, v := range replicationcontroller.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}
	return &js

}

func (j *ReplicationController) ObjectStatus() resource.ObjectStatus {
	return j.GetStatus()
}
func (j *ReplicationController) GetStatus() *Status {
	runtime, err := j.GetRuntime()
	if err != nil {
		js := Status{ObjectMeta: j.ObjectMeta}
		js.ContainerSpecs = make([]pk.ContainerSpec, 0)
		js.Images = make([]string, 0)
		js.PodStatus = make([]pk.Status, 0)
		js.Labels = make(map[string]string)
		js.Annotations = make(map[string]string)
		js.Selectors = make(map[string]string)
		js.Reason = err.Error()
		return &js
	}
	replicationcontroller := runtime.ReplicationController
	js := Status{ObjectMeta: j.ObjectMeta, ReplicationControllerStatus: replicationcontroller.Status}
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.Labels = make(map[string]string)
	js.Annotations = make(map[string]string)
	js.Selectors = make(map[string]string)

	if js.CreateTime == 0 {
		js.CreateTime = runtime.ReplicationController.CreationTimestamp.Unix()
	}

	if replicationcontroller.Spec.Replicas != nil {
		js.Replicas = *replicationcontroller.Spec.Replicas
	}
	if replicationcontroller.Labels != nil {
		js.Labels = replicationcontroller.Labels
	}

	if replicationcontroller.Annotations != nil {
		js.Labels = replicationcontroller.Annotations
	}

	if replicationcontroller.Spec.Selector != nil {
		js.Selectors = replicationcontroller.Spec.Selector
	}

	if replicationcontroller.Spec.Replicas != nil {
		js.Desire = int(*replicationcontroller.Spec.Replicas)
	} else {
		js.Desire = 1
	}
	js.Current = int(replicationcontroller.Status.AvailableReplicas)
	js.Ready = int(replicationcontroller.Status.ReadyReplicas)
	js.Available = int(replicationcontroller.Status.AvailableReplicas)

	for _, v := range replicationcontroller.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}
	//	js := K8sReplicationControllerToReplicationControllerStatus(runtime.ReplicationController)
	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
	}
	info := j.Info()
	js.User = info.User
	js.Group = info.Group
	js.Workspace = info.Workspace
	js.PodsCount = *resource.GetPodsCount(runtime.Pods)

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)
	}

	return &js
}
func (j *ReplicationController) Scale(num int) error {

	jh, err := cluster.NewReplicationControllerHandler(j.Group, j.Workspace)
	if err != nil {
		return err
	}

	err = jh.Scale(j.Workspace, j.Name, int32(num))
	if err != nil {
		return err
	}

	return nil
}

func (j *ReplicationController) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewReplicationControllerHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(j.Workspace, j.Name)
}

func (j *ReplicationController) GetTemplate() (string, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.ReplicationController)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: v1\nkind: ReplicationController"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil
}

func (p *ReplicationController) GetServices() ([]*corev1.Service, error) {
	ph, err := cluster.NewReplicationControllerHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.GetServices(p.Workspace, p.Name)
}

func (s *ReplicationController) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func (s *ReplicationController) GetRuntimeObjectCopy() (*corev1.ReplicationController, error) {
	r, err := s.GetRuntime()
	if err != nil {
		return nil, err
	}

	newobj := r.ReplicationController.DeepCopy()

	return newobj, nil
}

func InitReplicationControllerController(be backend.BackendHandler) (resource.ObjectController, error) {
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
	return rm, nil
}
