package secret

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

	yaml "gopkg.in/yaml.v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
)

var (
	rm         *SecretManager
	Controller resource.ObjectController
)

type SecretInterface interface {
	Info() *Secret
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
	GetReferenceObjects() ([]resource.ObjectReference, error)
}

type SecretManager struct {
	Groups map[string]SecretGroup `json:"groups"`
	locker sync.Mutex
}

type SecretGroup struct {
	Workspaces map[string]SecretWorkspace `json:"Workspaces"`
}

type SecretWorkspace struct {
	Secrets map[string]Secret `json:"secrets"`
}

type Runtime struct {
	*corev1.Secret
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Secret相关的K8s资源是否仍然存在
//在Secret构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Secret的信息
type Secret struct {
	resource.ObjectMeta
	Cluster string `json:"cluster"`
}

func GetSecretInterface(obj resource.Object) (SecretInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Secret)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *SecretManager) Lock() {
	p.locker.Lock()
}
func (p *SecretManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *SecretManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Secret{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *SecretManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Secret)
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
		_, ok = workspace.Secrets[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Secrets[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *SecretManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *SecretManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group SecretGroup
	group.Workspaces = make(map[string]SecretWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *SecretManager) ListGroups() []string {
	p.Lock()
	defer p.Unlock()
	gs := make([]string, 0)
	for k, _ := range p.Groups {
		gs = append(gs, k)
	}
	return nil
}

func (p *SecretManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Secret
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *SecretManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws SecretWorkspace
	ws.Secrets = make(map[string]Secret)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g

	//因为工作区事件的监听和集群的resource informers的监听是异步的,因此
	//工作区映射的命名空间实际创建时像sa/secret的资源会立即被创建,而且被resource informers已经
	//监听到,但是工作区事件因为延时的问题,导致没有把工作区告知informer controller.
	//这样informer controller认为该命名空间的资源的事件为可忽略的事件,从而忽略了资源的创建事件
	//从而导致工作区中缺失了该资源
	//因此在添加工作区时,获取一遍资源,更新到secret中
	ph, err := cluster.NewSecretHandler(groupName, workspaceName)
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

func (p *SecretManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *SecretManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *SecretManager) get(groupName, workspaceName, resourceName string) (*Secret, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	secret, ok := workspace.Secrets[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &secret, nil
}

func (p *SecretManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *SecretManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}
func (p *SecretManager) ListGroupWorkspaceObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.Secrets {
		t := workspace.Secrets[k]
		pis = append(pis, &t)
	}

	return pis, nil
}

func (p *SecretManager) ListGroupObject(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Secrets {
			t := v.Secrets[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *SecretManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()

	ph, err := cluster.NewSecretHandler(groupName, workspaceName)
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

	var obj corev1.Secret
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

	var cp Secret
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Comment = opt.Comment
	cp.Workspace = workspaceName
	cp.Kind = resourceKind
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
func (p *SecretManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Secrets, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *SecretManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()
	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	res, err := p.get(group, workspace, resourceName)
	if err != nil {
		return log.DebugPrint(err)
	}
	ph, err := cluster.NewSecretHandler(group, workspace)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return log.DebugPrint(err)
		}
	}

	if res.MemoryOnly {

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

func (p *SecretManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr corev1.Secret
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""
	if newr.Annotations == nil {
		newr.Annotations = make(map[string]string)
	}
	if !res.MemoryOnly {
		newr.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue
	}

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, resource name not match")
	}

	ph, err := cluster.NewSecretHandler(groupName, workspaceName)
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
func (secret *Secret) Info() *Secret {
	return secret
}

func (s *Secret) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewSecretHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{Secret: svc}, nil
}

func (s *Secret) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.Secret)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: v1\nkind: Secret"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	Reason     string            `json:"reason"`
	Type       string            `json:"type"`
	Data       map[string][]byte `json:"data"`
	StringData map[string]string `json:"stringdata"`
	DataString string            `json:"datastring"`
}

func (s *Secret) ObjectStatus() resource.ObjectStatus {
	return s.GetStatus()
}
func (s *Secret) GetStatus() *Status {
	runtime, err := s.GetRuntime()

	js := Status{ObjectMeta: s.ObjectMeta}
	js.Data = make(map[string][]byte)
	js.StringData = make(map[string]string)
	if err != nil {
		js.Reason = err.Error()
		return &js
	}

	if js.CreateTime == 0 {
		js.CreateTime = runtime.Secret.CreationTimestamp.Unix()
	}

	js.Type = string(runtime.Type)
	js.StringData = runtime.StringData
	js.Data = runtime.Data

	bc, err := yaml.Marshal(runtime.Data)
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	js.DataString = string(bc)
	return &js
}

func (s *Secret) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}

func (s *Secret) GetReferenceObjects() ([]resource.ObjectReference, error) {
	ph, err := cluster.NewSecretHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	apiors, err := ph.GetReferenceResources(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}

	ors := make([]resource.ObjectReference, 0)
	for _, v := range apiors {
		var or resource.ObjectReference
		or.ObjectReference = v
		or.Namespace = s.Workspace
		or.Group = s.Group
		ors = append(ors, or)

	}
	return ors, nil
}

func (s *Secret) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitSecretController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &SecretManager{}
	rm.Groups = make(map[string]SecretGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for k, v := range rs {
		var group SecretGroup
		group.Workspaces = make(map[string]SecretWorkspace)
		for i, j := range v.Workspaces {
			var workspace SecretWorkspace
			workspace.Secrets = make(map[string]Secret)
			for m, n := range j.Resources {
				var secret Secret
				err := json.Unmarshal([]byte(n), &secret)
				if err != nil {
					return nil, fmt.Errorf("init secret manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Secrets[m] = secret
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil

}
