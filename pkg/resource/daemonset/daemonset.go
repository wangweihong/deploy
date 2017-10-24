package daemonset

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/util"

	corev1 "k8s.io/client-go/pkg/api/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm *DaemonSetManager
	/* = &DaemonSetManager{
		Groups: make(map[string]DaemonSetGroup),
		locker: sync.Mutex{},
	}
	*/
	Controller DaemonSetController

	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrGroupNotFound     = fmt.Errorf("group not found")
)

type DaemonSetController interface {
	Create(group, workspace string, data []byte, opt resource.CreateOption) error
	Delete(group, workspace, daemonset string, opt resource.DeleteOption) error
	Get(group, workspace, daemonset string) (DaemonSetInterface, error)
	Update(group, workspace, resource string, newdata []byte) error
	List(group, workspace string) ([]DaemonSetInterface, error)
	ListGroup(groupName string) ([]DaemonSetInterface, error)
}

type DaemonSetInterface interface {
	Info() *DaemonSet
	GetRuntime() (*Runtime, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
	GetTemplate() (string, error)
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
	memoryOnly bool
}

//注意这里没锁
func (p *DaemonSetManager) get(groupName, workspaceName, resourceName string) (*DaemonSet, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, ErrWorkspaceNotFound
	}

	daemonset, ok := workspace.DaemonSets[resourceName]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return &daemonset, nil
}

func (p *DaemonSetManager) Get(group, workspace, resourceName string) (DaemonSetInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *DaemonSetManager) List(groupName, workspaceName string) ([]DaemonSetInterface, error) {

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

	pis := make([]DaemonSetInterface, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for k := range workspace.DaemonSets {
		t := workspace.DaemonSets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *DaemonSetManager) ListGroup(groupName string) ([]DaemonSetInterface, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", ErrGroupNotFound, groupName)
	}

	pis := make([]DaemonSetInterface, 0)
	for _, v := range group.Workspaces {
		for k := range v.DaemonSets {
			t := v.DaemonSets[k]
			pis = append(pis, &t)
		}
	}
	return pis, nil
}

func (p *DaemonSetManager) Create(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

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

	var svc extensionsv1beta1.DaemonSet
	err = json.Unmarshal(exts[0].Raw, &svc)
	if err != nil {
		return log.DebugPrint(err)
	}

	if svc.Kind != "DaemonSet" {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	svc.ResourceVersion = ""

	var cp DaemonSet
	cp.CreateTime = time.Now().Unix()
	cp.Name = svc.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
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

	err = ph.Create(workspaceName, &svc)
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
		return ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return ErrWorkspaceNotFound
	}

	delete(workspace.DaemonSets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *DaemonSetManager) Delete(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewDaemonSetHandler(group, workspace)
	if err != nil {
		return log.DebugPrint(err)
	}
	res, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if res.memoryOnly {

		//触发集群控制器来删除内存中的数据
		err = ph.Delete(workspace, resourceName)
		if err != nil {
			return log.DebugPrint(err)
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
		if !opt.DontCallApp && res.App != "" {
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

func (p *DaemonSetManager) Update(groupName, workspaceName string, resourceName string, data []byte) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	_, err := p.get(groupName, workspaceName, resourceName)
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
	err = ph.Update(workspaceName, &newr)
	if err != nil {
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
	UpdateStrategy string            `json:"updatestrategy"`
	//	Pods       []string `json:"pods"`
	PodStatus      []pk.Status        `json:"podstatus"`
	ContainerSpecs []pk.ContainerSpec `json:"containerspec"`
	extensionsv1beta1.DaemonSetStatus
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

func InitDaemonSetController(be backend.BackendHandler) (DaemonSetController, error) {
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
	log.DebugPrint(rm)
	return rm, nil

}
