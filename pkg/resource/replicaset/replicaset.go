package replicaset

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
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm         *ReplicaSetManager
	Controller resource.ObjectController
)

type ReplicaSetInterface interface {
	Info() *ReplicaSet
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	Scale(num int) error
	Event() ([]corev1.Event, error)
	GetTemplate() (string, error)
	GetServices() ([]*corev1.Service, error)

	//	Runtime()
}

type ReplicaSetManager struct {
	Groups map[string]ReplicaSetGroup `json:"groups"`
	locker sync.Mutex
}

type ReplicaSetGroup struct {
	Workspaces map[string]ReplicaSetWorkspace `json:"Workspaces"`
}

type ReplicaSetWorkspace struct {
	ReplicaSets map[string]ReplicaSet `json:"replicasets"`
}

type Runtime struct {
	ReplicaSet *extensionsv1beta1.ReplicaSet
	Pods       []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记ReplicaSet相关的K8s资源是否仍然存在
//在ReplicaSet构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新ReplicaSet的信息
type ReplicaSet struct {
	resource.ObjectMeta
}

func GetReplicaSetInterface(obj resource.Object) (ReplicaSetInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*ReplicaSet)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *ReplicaSetManager) Lock() {
	p.locker.Lock()
}

func (p *ReplicaSetManager) Unlock() {
	p.locker.Unlock()
}

func (p *ReplicaSetManager) Kind() string {
	return resourceKind
}

//仅仅用于基于内存的对象的创建
func (p *ReplicaSetManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := ReplicaSet{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *ReplicaSetManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*ReplicaSet)
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
		_, ok = workspace.ReplicaSets[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.ReplicaSets[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *ReplicaSetManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *ReplicaSetManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group ReplicaSetGroup
	group.Workspaces = make(map[string]ReplicaSetWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *ReplicaSetManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *ReplicaSetManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res ReplicaSet
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *ReplicaSetManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws ReplicaSetWorkspace
	ws.ReplicaSets = make(map[string]ReplicaSet)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewReplicaSetHandler(groupName, workspaceName)
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

func (p *ReplicaSetManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *ReplicaSetManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *ReplicaSetManager) get(groupName, workspaceName, replicasetName string) (*ReplicaSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	replicaset, ok := workspace.ReplicaSets[replicasetName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &replicaset, nil
}

func (p *ReplicaSetManager) GetObject(group, workspace, replicasetName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, replicasetName)
}
func (p *ReplicaSetManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *ReplicaSetManager) ListGroupObject(groupName string) ([]resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)
	for _, v := range group.Workspaces {
		for k := range v.ReplicaSets {
			t := v.ReplicaSets[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *ReplicaSetManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.ReplicaSets {
		t := workspace.ReplicaSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *ReplicaSetManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewReplicaSetHandler(groupName, workspaceName)
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

	var obj extensionsv1beta1.ReplicaSet
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

	var cp ReplicaSet
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Kind = resourceKind
	cp.Comment = opt.Comment
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
func (p *ReplicaSetManager) delete(groupName, workspaceName, replicasetName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.ReplicaSets, replicasetName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *ReplicaSetManager) DeleteObject(group, workspace, replicasetName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	if opt.MemoryOnly {
		return p.delete(group, workspace, replicasetName)
	}

	ph, err := cluster.NewReplicaSetHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}

	res, err := p.get(group, workspace, replicasetName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.MemoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, replicasetName)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return log.DebugPrint(err)
			}
		}
		return nil
		//TODO:ufleet创建的数据
	} else {
		be := backend.NewBackendHandler()
		err := be.DeleteResource(backendKind, group, workspace, replicasetName)
		if err != nil {
			return log.DebugPrint(err)
		}
		err = ph.Delete(workspace, replicasetName)
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

func (p *ReplicaSetManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.ReplicaSet
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

	ph, err := cluster.NewReplicaSetHandler(groupName, workspaceName)
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
func (j *ReplicaSet) Info() *ReplicaSet {
	return j
}

func (j *ReplicaSet) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewReplicaSetHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	replicaset, err := ph.Get(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	pods, err := ph.GetPods(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	var runtime Runtime
	runtime.Pods = pods
	runtime.ReplicaSet = replicaset
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
	Desire          int                       `json:"desire"`
	Current         int                       `json:"current"`
	Available       int                       `json:"available"`
	Ready           int                       `json:"ready"`
	Labels          map[string]string         `json:"labels"`
	Annotations     map[string]string         `json:"annotations"`
	Selectors       map[string]string         `json:"selector"`
	Reason          string                    `json:"reason"`
	OwnerReferences []resource.OwnerReference `json:"ownerreferences"`

	//	Pods       []string `json:"pods"`
	ContainerSpecs []pk.ContainerSpec `json:"containerspec"`
	PodStatus      []pk.Status        `json:"podstatus"`
	PodsCount      resource.PodsCount `json:"podscount"`
	extensionsv1beta1.ReplicaSetStatus
}

func (j *ReplicaSet) ObjectStatus() resource.ObjectStatus {
	return j.GetStatus()
}
func (j *ReplicaSet) GetStatus() *Status {
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
	replicaset := runtime.ReplicaSet
	js := Status{ObjectMeta: j.ObjectMeta, ReplicaSetStatus: replicaset.Status}
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.Labels = make(map[string]string)
	js.Annotations = make(map[string]string)
	js.Selectors = make(map[string]string)

	if js.CreateTime == 0 {
		js.CreateTime = runtime.ReplicaSet.CreationTimestamp.Unix()
	}
	js.OwnerReferences = make([]resource.OwnerReference, 0)

	if runtime.ReplicaSet.OwnerReferences != nil {

		//var cr resource.ObjectReference
		for k := range replicaset.OwnerReferences {
			cr := resource.OwnerReference{OwnerReference: replicaset.OwnerReferences[k], Group: j.Group, Namespace: j.Workspace}
			js.OwnerReferences = append(js.OwnerReferences, cr)
		}

	}

	if replicaset.Spec.Replicas != nil {
		js.Replicas = *replicaset.Spec.Replicas
	}
	if replicaset.Labels != nil {
		js.Labels = replicaset.Labels
	}

	if replicaset.Annotations != nil {
		js.Annotations = replicaset.Annotations
	}

	if replicaset.Spec.Selector != nil {
		js.Selectors = replicaset.Spec.Selector.MatchLabels
	}

	if replicaset.Spec.Replicas != nil {
		js.Desire = int(*replicaset.Spec.Replicas)
	} else {
		js.Desire = 1
	}
	js.Current = int(replicaset.Status.AvailableReplicas)
	js.Ready = int(replicaset.Status.ReadyReplicas)
	js.Available = int(replicaset.Status.AvailableReplicas)

	for _, v := range replicaset.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}
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

	js.PodsCount = *resource.GetPodsCount(runtime.Pods)
	return &js
}

func (j *ReplicaSet) Scale(num int) error {

	jh, err := cluster.NewReplicaSetHandler(j.Group, j.Workspace)
	if err != nil {
		return err
	}

	err = jh.Scale(j.Workspace, j.Name, int32(num))
	if err != nil {
		return err
	}

	return nil
}

func (j *ReplicaSet) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewReplicaSetHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(j.Workspace, j.Name)
}

func (j *ReplicaSet) GetTemplate() (string, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.ReplicaSet)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: v1\nkind: ReplicaSet"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil
}

func (p *ReplicaSet) GetServices() ([]*corev1.Service, error) {
	ph, err := cluster.NewReplicaSetHandler(p.Group, p.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.GetServices(p.Workspace, p.Name)
}

func (s *ReplicaSet) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}
func InitReplicaSetController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &ReplicaSetManager{}
	rm.Groups = make(map[string]ReplicaSetGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group ReplicaSetGroup
		group.Workspaces = make(map[string]ReplicaSetWorkspace)
		for i, j := range v.Workspaces {
			var workspace ReplicaSetWorkspace
			workspace.ReplicaSets = make(map[string]ReplicaSet)
			for m, n := range j.Resources {
				var replicaset ReplicaSet
				err := json.Unmarshal([]byte(n), &replicaset)
				if err != nil {
					return nil, fmt.Errorf("init replicaset manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.ReplicaSets[m] = replicaset
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil
}
