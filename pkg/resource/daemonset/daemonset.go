package daemonset

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

	corev1 "k8s.io/client-go/pkg/api/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm         *DaemonSetManager
	Controller resource.ObjectController
)

type DaemonSetInterface interface {
	Info() *DaemonSet
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	ObjectStatus() resource.ObjectStatus
	Event() ([]corev1.Event, error)
	GetTemplate() (string, error)
	GetRevisionsAndDescribe() (map[int64]string, error)
	Rollback(revision int64) (*string, error)
}

type DaemonSetManager struct {
	Groups map[string]DaemonSetGroup `json:"groups"`
	locker sync.Mutex
}

type DaemonSetGroup struct {
	Workspaces map[string]DaemonSetWorkspace `json:"Workspaces"`
}

type DaemonSetWorkspace struct {
	DaemonSets map[string]DaemonSet `json:"daemonsets"`
}

type Runtime struct {
	DaemonSet *extensionsv1beta1.DaemonSet
	Pods      []*corev1.Pod
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记DaemonSet相关的K8s资源是否仍然存在
//在DaemonSet构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新DaemonSet的信息
type DaemonSet struct {
	resource.ObjectMeta
}

func (p *DaemonSetManager) Lock() {
	p.locker.Lock()
}
func (p *DaemonSetManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *DaemonSetManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := DaemonSet{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *DaemonSetManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*DaemonSet)
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
		_, ok = workspace.DaemonSets[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}

	}
	workspace.DaemonSets[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *DaemonSetManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *DaemonSetManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group DaemonSetGroup
	group.Workspaces = make(map[string]DaemonSetWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *DaemonSetManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res DaemonSet
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *DaemonSetManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws DaemonSetWorkspace
	ws.DaemonSets = make(map[string]DaemonSet)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *DaemonSetManager) DeleteWorkspace(groupName string, workspaceName string) error {
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
func (p *DaemonSetManager) get(groupName, workspaceName, resourceName string) (*DaemonSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	daemonset, ok := workspace.DaemonSets[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &daemonset, nil
}

func (p *DaemonSetManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {
	return p.get(groupName, workspaceName, resourceName)
}

func (p *DaemonSetManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *DaemonSetManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.DaemonSets {
		t := workspace.DaemonSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *DaemonSetManager) ListGroup(groupName string) ([]resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)
	for _, v := range group.Workspaces {
		for k := range v.DaemonSets {
			t := v.DaemonSets[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *DaemonSetManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewDaemonSetHandler(groupName, workspaceName)
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

	var obj extensionsv1beta1.DaemonSet
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	obj.ResourceVersion = ""
	obj.Annotations = make(map[string]string)
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp DaemonSet
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
	cp.CreateTime = time.Now().Unix()
	cp.Kind = resourceKind
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
func (p *DaemonSetManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.DaemonSets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DaemonSetManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewDaemonSetHandler(group, workspace)
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

func (p *DaemonSetManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.DaemonSet
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewDaemonSetHandler(groupName, workspaceName)
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

func (daemonset *DaemonSet) Info() *DaemonSet {
	return daemonset
}

func (j *DaemonSet) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewDaemonSetHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	daemonset, err := ph.Get(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	pods, err := ph.GetPods(j.Workspace, j.Name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	var runtime Runtime
	runtime.Pods = pods
	runtime.DaemonSet = daemonset
	return &runtime, nil
}

type Status struct {
	resource.ObjectMeta
	Images         []string          `json:"images"`
	Containers     []string          `json:"containers"`
	PodNum         int               `json:"podnum"`
	ClusterIP      string            `json:"clusterip"`
	Labels         map[string]string `json:"labels"`
	Annotatiosn    map[string]string `json:"annotations"`
	Selectors      map[string]string `json:"selectors"`
	Reason         string            `json:"reason"`
	Revision       int64             `json:"revision"`
	UpdateStrategy string            `json:"updatestrategy"`
	//	Pods       []string `json:"pods"`
	PodStatus      []pk.Status        `json:"podstatus"`
	ContainerSpecs []pk.ContainerSpec `json:"containerspec"`
	extensionsv1beta1.DaemonSetStatus
}

func (j *DaemonSet) ObjectStatus() resource.ObjectStatus {
	return j.GetStatus()

}

func (j *DaemonSet) GetStatus() *Status {
	runtime, err := j.GetRuntime()
	if err != nil {
		js := Status{ObjectMeta: j.ObjectMeta}
		js.Images = make([]string, 0)
		js.PodStatus = make([]pk.Status, 0)
		js.ContainerSpecs = make([]pk.ContainerSpec, 0)
		js.Labels = make(map[string]string)
		js.Annotatiosn = make(map[string]string)
		js.Selectors = make(map[string]string)
		js.Reason = err.Error()
		return &js
	}

	daemonset := runtime.DaemonSet
	js := Status{ObjectMeta: j.ObjectMeta, DaemonSetStatus: runtime.DaemonSet.Status}
	js.Images = make([]string, 0)
	js.PodStatus = make([]pk.Status, 0)
	js.ContainerSpecs = make([]pk.ContainerSpec, 0)
	js.Labels = make(map[string]string)
	js.Annotatiosn = make(map[string]string)
	js.Selectors = make(map[string]string)

	if js.CreateTime == 0 {
		js.CreateTime = daemonset.CreationTimestamp.Unix()
	}

	if daemonset.Labels != nil {
		js.Labels = daemonset.Labels
	}

	if daemonset.Annotations != nil {
		js.Labels = daemonset.Annotations
	}

	if daemonset.Spec.Selector != nil {
		js.Selectors = daemonset.Spec.Selector.MatchLabels
	}
	//	js := K8sDaemonSetToDaemonSetStatus(runtime.DaemonSet)
	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
	}

	js.Revision = runtime.DaemonSet.Generation

	js.UpdateStrategy = string(daemonset.Spec.UpdateStrategy.Type)
	for _, v := range daemonset.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
		js.Images = append(js.Images, v.Image)
		js.ContainerSpecs = append(js.ContainerSpecs, *pk.K8sContainerSpecTran(&v))
	}

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)
	}

	return &js
}

func (j *DaemonSet) Event() ([]corev1.Event, error) {
	ph, err := cluster.NewDaemonSetHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return ph.Event(j.Workspace, j.Name)
}

func (j *DaemonSet) GetTemplate() (string, error) {
	runtime, err := j.GetRuntime()
	if err != nil {
		return "", log.DebugPrint(err)
	}

	t, err := util.GetYamlTemplateFromObject(runtime.DaemonSet)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: extensions/v1beta1\nkind: DaemonSet"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

	return *t, nil
}

func (j *DaemonSet) GetRevisionsAndDescribe() (map[int64]string, error) {
	ph, err := cluster.NewDaemonSetHandler(j.Group, j.Workspace)
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

func (j *DaemonSet) Rollback(revision int64) (*string, error) {
	ph, err := cluster.NewDaemonSetHandler(j.Group, j.Workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	return ph.Rollback(j.Workspace, j.Name, revision)

}
func (s *DaemonSet) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}
func GetDaemonSetInterface(obj resource.Object) (DaemonSetInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*DaemonSet)
	if !ok {
		return nil, fmt.Errorf("resource object is not DaemonSet type")
	}

	return ri, nil

}

func InitDaemonSetController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &DaemonSetManager{}
	rm.Groups = make(map[string]DaemonSetGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group DaemonSetGroup
		group.Workspaces = make(map[string]DaemonSetWorkspace)
		for i, j := range v.Workspaces {
			var workspace DaemonSetWorkspace
			workspace.DaemonSets = make(map[string]DaemonSet)
			for m, n := range j.Resources {
				var daemonset DaemonSet
				err := json.Unmarshal([]byte(n), &daemonset)
				if err != nil {
					return nil, fmt.Errorf("init daemonset manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.DaemonSets[m] = daemonset
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
