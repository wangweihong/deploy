package ingress

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	rm         *IngressManager
	Controller resource.ObjectController
)

type IngressInterface interface {
	Info() *Ingress
	GetRuntime() (*Runtime, error)
	GetTemplate() (string, error)
	GetStatus() *Status
	Event() ([]corev1.Event, error)
}

type IngressManager struct {
	Groups map[string]IngressGroup `json:"groups"`
	locker sync.Mutex
}

type IngressGroup struct {
	Workspaces map[string]IngressWorkspace `json:"Workspaces"`
}

type IngressWorkspace struct {
	Ingresss map[string]Ingress `json:"ingresss"`
}

type Runtime struct {
	Ingress *extensionsv1beta1.Ingress
}

//TODO:是否可以添加一个特定的只存于内存的标记位
//用于标记Ingress相关的K8s资源是否仍然存在
//在Ingress构建到内存的时候,就开始绑定K8s资源,
//可以根据事件及时更新Ingress的信息
type Ingress struct {
	resource.ObjectMeta
}

func GetIngressInterface(obj resource.Object) (IngressInterface, error) {
	if obj == nil {
		return nil, fmt.Errorf("resource object is nil")
	}

	ri, ok := obj.(*Ingress)
	if !ok {
		return nil, fmt.Errorf("resource object is not configmap type")
	}

	return ri, nil
}

func (p *IngressManager) Lock() {
	p.locker.Lock()
}
func (p *IngressManager) Unlock() {
	p.locker.Unlock()
}

//仅仅用于基于内存的对象的创建
func (p *IngressManager) NewObject(meta resource.ObjectMeta) error {

	if strings.TrimSpace(meta.Group) == "" ||
		strings.TrimSpace(meta.Workspace) == "" ||
		strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("Invalid object data")
	}

	cp := Ingress{ObjectMeta: meta}
	cp.MemoryOnly = true

	err := p.fillObjectToManager(&cp, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *IngressManager) fillObjectToManager(meta resource.Object, force bool) error {

	cm, ok := meta.(*Ingress)
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
		_, ok = workspace.Ingresss[cm.Name]
		if ok {
			return resource.ErrResourceExists
		}
	}

	workspace.Ingresss[cm.Name] = *cm
	group.Workspaces[cm.Workspace] = workspace
	p.Groups[cm.Group] = group
	return nil

}

func (p *IngressManager) DeleteGroup(groupName string) error {
	_, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}

	delete(p.Groups, groupName)
	return nil
}

func (p *IngressManager) AddGroup(groupName string) error {
	p.Lock()
	defer p.Unlock()
	_, ok := p.Groups[groupName]
	if ok {
		return resource.ErrGroupExists
	}
	var group IngressGroup
	group.Workspaces = make(map[string]IngressWorkspace)
	p.Groups[groupName] = group
	return nil
}

func (p *IngressManager) AddObjectFromBytes(data []byte, force bool) error {
	p.Lock()
	defer p.Unlock()
	var res Ingress
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}
	err = p.fillObjectToManager(&res, force)
	return err

}

func (p *IngressManager) AddWorkspace(groupName string, workspaceName string) error {
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

	var ws IngressWorkspace
	ws.Ingresss = make(map[string]Ingress)
	g.Workspaces[workspaceName] = ws
	p.Groups[groupName] = g
	return nil

}

func (p *IngressManager) DeleteWorkspace(groupName string, workspaceName string) error {
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

func (p *IngressManager) GetObjectWithoutLock(groupName, workspaceName, resourceName string) (resource.Object, error) {

	return p.get(groupName, workspaceName, resourceName)
}

//注意这里没锁
func (p *IngressManager) get(groupName, workspaceName, resourceName string) (*Ingress, error) {

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, resource.ErrGroupNotFound
	}

	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return nil, resource.ErrWorkspaceNotFound
	}

	ingress, ok := workspace.Ingresss[resourceName]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	return &ingress, nil
}

func (p *IngressManager) GetObject(group, workspace, resourceName string) (resource.Object, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.get(group, workspace, resourceName)
}

func (p *IngressManager) GetObjectTemplate(group, workspace, resourceName string) (string, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	s, err := p.get(group, workspace, resourceName)
	if err != nil {
		return "", err
	}
	return s.GetTemplate()
}

func (p *IngressManager) ListObject(groupName, workspaceName string) ([]resource.Object, error) {

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
	for k := range workspace.Ingresss {
		t := workspace.Ingresss[k]
		pis = append(pis, &t)
	}

	return pis, nil
}
func (p *IngressManager) ListGroup(groupName string) ([]resource.Object, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	group, ok := p.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%v:%v", resource.ErrGroupNotFound, groupName)
	}

	pis := make([]resource.Object, 0)

	//不能够直接使用k,v来赋值,会出现值都是同一个的问题
	for _, v := range group.Workspaces {
		for k := range v.Ingresss {
			t := v.Ingresss[k]
			pis = append(pis, &t)
		}
	}

	return pis, nil
}

func (p *IngressManager) CreateObject(groupName, workspaceName string, data []byte, opt resource.CreateOption) error {

	p.locker.Lock()
	defer p.locker.Unlock()
	ph, err := cluster.NewIngressHandler(groupName, workspaceName)
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

	var obj extensionsv1beta1.Ingress
	obj.Annotations = make(map[string]string)
	err = json.Unmarshal(exts[0].Raw, &obj)
	if err != nil {
		return log.DebugPrint(err)
	}

	if obj.Kind != resourceKind {
		return log.DebugPrint("must and  offer one resource json/yaml data")
	}
	obj.ResourceVersion = ""
	obj.Annotations[sign.SignFromUfleetKey] = sign.SignFromUfleetValue

	var cp Ingress
	cp.CreateTime = time.Now().Unix()
	cp.Name = obj.Name
	cp.Workspace = workspaceName
	cp.Group = groupName
	cp.Template = string(data)
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
func (p *IngressManager) delete(groupName, workspaceName, resourceName string) error {
	group, ok := p.Groups[groupName]
	if !ok {
		return resource.ErrGroupNotFound
	}
	workspace, ok := group.Workspaces[workspaceName]
	if !ok {
		return resource.ErrWorkspaceNotFound
	}

	delete(workspace.Ingresss, resourceName)
	group.Workspaces[workspaceName] = workspace
	p.Groups[groupName] = group
	return nil
}

func (p *IngressManager) DeleteObject(group, workspace, resourceName string, opt resource.DeleteOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	if opt.MemoryOnly {
		return p.delete(group, workspace, resourceName)
	}

	ph, err := cluster.NewIngressHandler(group, workspace)
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

func (p *IngressManager) UpdateObject(groupName, workspaceName string, resourceName string, data []byte, opt resource.UpdateOption) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	res, err := p.get(groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}

	//说明是主动创建的..
	var newr extensionsv1beta1.Ingress
	err = util.GetObjectFromYamlTemplate(data, &newr)
	if err != nil {
		return log.DebugPrint(err)
	}
	//
	newr.ResourceVersion = ""

	if newr.Name != resourceName {
		return fmt.Errorf("invalid update data, name not match")
	}

	ph, err := cluster.NewIngressHandler(groupName, workspaceName)
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
func (ingress *Ingress) Info() *Ingress {
	return ingress
}
func (s *Ingress) GetRuntime() (*Runtime, error) {
	ph, err := cluster.NewIngressHandler(s.Group, s.Workspace)
	if err != nil {
		return nil, err
	}

	svc, err := ph.Get(s.Workspace, s.Name)
	if err != nil {
		return nil, err
	}
	return &Runtime{Ingress: svc}, nil
}

func (s *Ingress) GetTemplate() (string, error) {
	runtime, err := s.GetRuntime()
	if err != nil {
		return "", err
	}
	t, err := util.GetYamlTemplateFromObject(runtime.Ingress)
	if err != nil {
		return "", log.DebugPrint(err)
	}

	prefix := "apiVersion: extensions/v1beta1\nkind: Ingress"
	*t = fmt.Sprintf("%v\n%v", prefix, *t)
	return *t, nil

}

type Status struct {
	resource.ObjectMeta
	Hosts         []string `json:"hosts"`
	Ports         []int    `json:"ports"`
	IngressSpec   `json:"ingressspec"`
	IngressStatus `json:"ingressstatus"`
	Reason        string            `json:"reason"`
	Labels        map[string]string `json:"labels"`
}

type IngressSpec struct {
	Backend *extensionsv1beta1.IngressBackend `json:"backend"`
	TLS     []IngressTLS                      `json:"tls"`
	Rules   []IngressRule                     `json:"rules"`
}
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

type IngressRule struct {
	Host string                                  `json:"host"`
	HTTP *extensionsv1beta1.HTTPIngressRuleValue `json:"http"`
}

type IngressStatus struct {
	LoadBalancer corev1.LoadBalancerStatus `json:"loadbalancer"`
}

func (s *Ingress) ObjectStatus() resource.ObjectStatus {
	return s.GetStatus()
}

func (s *Ingress) GetStatus() *Status {
	js := Status{ObjectMeta: s.ObjectMeta}
	js.Hosts = make([]string, 0)
	js.Ports = make([]int, 0)
	js.IngressSpec.Rules = make([]IngressRule, 0)
	js.IngressSpec.TLS = make([]IngressTLS, 0)
	js.IngressStatus.LoadBalancer.Ingress = make([]corev1.LoadBalancerIngress, 0)
	js.Labels = make(map[string]string)

	runtime, err := s.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return &js
	}
	if js.CreateTime == 0 {
		js.CreateTime = runtime.Ingress.CreationTimestamp.Unix()
	}

	res := runtime.Ingress
	if res.Spec.Backend != nil {
		js.IngressSpec.Backend = res.Spec.Backend
	}
	//js.IngressSpec.Rules = append(js.IngressSpec.Rules, res.Spec.Rules...)
	for _, v := range res.Spec.Rules {
		js.Hosts = append(js.Hosts, v.Host)
		var ir IngressRule
		ir.Host = v.Host
		ir.HTTP = v.HTTP
		js.IngressSpec.Rules = append(js.IngressSpec.Rules, ir)
	}
	if len(js.Hosts) == 0 {
		js.Hosts = append(js.Hosts, "*")
	}

	js.Ports = append(js.Ports, 80)
	for _, v := range res.Spec.TLS {
		var tls IngressTLS
		tls.Hosts = make([]string, 0)
		tls.SecretName = v.SecretName
		tls.Hosts = append(tls.Hosts, v.Hosts...)
		//	js.TLS = append(js.TLS, tls)
		js.IngressSpec.TLS = append(js.IngressSpec.TLS, tls)
	}

	if len(js.IngressSpec.TLS) != 0 {
		js.Ports = append(js.Ports, 443)
	}
	js.Labels = res.Labels

	js.IngressStatus.LoadBalancer.Ingress = append(js.IngressStatus.LoadBalancer.Ingress, res.Status.LoadBalancer.Ingress...)

	return &js
}
func (s *Ingress) Event() ([]corev1.Event, error) {
	e := make([]corev1.Event, 0)
	return e, nil
}
func (s *Ingress) Metadata() resource.ObjectMeta {
	return s.ObjectMeta
}

func InitIngressController(be backend.BackendHandler) (resource.ObjectController, error) {
	rm = &IngressManager{}
	rm.Groups = make(map[string]IngressGroup)
	rm.locker = sync.Mutex{}

	rs, err := be.GetResourceAllGroup(backendKind)
	if err != nil {
		return nil, err
	}

	for k, v := range rs {
		var group IngressGroup
		group.Workspaces = make(map[string]IngressWorkspace)
		for i, j := range v.Workspaces {
			var workspace IngressWorkspace
			workspace.Ingresss = make(map[string]Ingress)
			for m, n := range j.Resources {
				var ingress Ingress
				err := json.Unmarshal([]byte(n), &ingress)
				if err != nil {
					return nil, fmt.Errorf("init ingress manager fail for unmarshal \"%v\" for %v", string(n), err)
				}
				workspace.Ingresss[m] = ingress
			}
			group.Workspaces[i] = workspace
		}
		rm.Groups[k] = group
	}
	return rm, nil
}
