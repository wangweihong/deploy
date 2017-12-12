package resource

import (
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/pkg/api/v1"
)

const (
	DefaultAppBelong = "" //"standalone"

	PodReady       = "ready"
	PodTerminating = "terminating"
)

var (
	resourceToController = make(map[string]ObjectController)
	locker               = sync.Mutex{}

	//通知App 资源的时间
	ResourceEventChan    = make(chan ResourceEvent)
	ResourceActionDelete = "delete"

//	ResourceActionCreate = "create"
)

type ResourceEvent struct {
	Group     string
	Workspace string
	Resource  string
	Kind      string
	Action    string
	App       string
}

type CreateOption struct {
	App     *string //所属app
	User    string  //创建的用户
	Comment string  //注释
}
type ListOption struct{}
type DeleteOption struct {
	DontCallApp bool
	MemoryOnly  bool //只清除内存中的数据
}

type UpdateOption struct {
	Comment string //注释
}

//抽象,便于app使用
func RegisterResourceController(name string, cud ObjectController) error {
	locker.Lock()
	defer locker.Unlock()

	if _, ok := resourceToController[name]; ok {
		return fmt.Errorf("resource %v has register to RCUD", name)
	}

	resourceToController[name] = cud
	return nil
}

func GetResourceController(name string) (ObjectController, error) {
	locker.Lock()
	defer locker.Unlock()

	cud, ok := resourceToController[name]
	if !ok {
		return nil, fmt.Errorf("resource %v doesn't register ", name)
	}
	return cud, nil

}

type ObjectMeta struct {
	Name      string `json:"name"`
	Workspace string `json:"workspace"`
	Group     string `json:"group"`
	App       string `json:"app"`
	User      string `json:"user"`

	Kind       string `json:"kind"`
	Template   string `json:"template"`
	CreateTime int64  `json:"createtime"`
	Comment    string `json:"comment"`
	MemoryOnly bool   `json:"memoryonly"`
}

type Object interface {
	Metadata() ObjectMeta
	ObjectStatus() ObjectStatus
}

//所有Object的Status都应该嵌套这个
type ObjectStatus interface {
}

type Locker interface {
	Lock()
	Unlock()
}

type ObjectController interface {
	Locker
	Kind() string
	NewObject(meta ObjectMeta) error
	GetObjectWithoutLock(group, workspace, name string) (Object, error)
	DeleteGroup(group string) error
	DeleteWorkspace(groupName string, workspaceName string) error
	AddGroup(group string) error
	ListGroups() []string
	AddWorkspace(group, workspace string) error
	AddObjectFromBytes(data []byte, force bool) error

	GetObjectTemplate(group, workspace, resourceName string) (string, error)
	CreateObject(group, workspace string, data []byte, opt CreateOption) error
	DeleteObject(group, workspace, resource string, opt DeleteOption) error
	GetObject(group, workspace, resource string) (Object, error)
	UpdateObject(group, workspace, resource string, newdata []byte, opt UpdateOption) error
	ListGroupWorkspaceObject(group, workspace string) ([]Object, error)
	ListGroupObject(group string) ([]Object, error)
}

//env

type EnvVar struct {
}

type ObjectReference struct {
	corev1.ObjectReference
	Group     string `group`
	Namespace string `json:"workspace"`
}

type OwnerReference struct {
	metav1.OwnerReference
	Group     string `group`
	Namespace string `json:"workspace"`
}

type PodsCount struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Succeeded int `json:"successed"`
	Failed    int `json:"failed"`
	Ready     int `json:"ready"`
	Unknown   int `json:"unknown"`
}

func GetPodsCount(pis interface{}) *PodsCount {

	c := &PodsCount{}
	if pods, ok := pis.([]*corev1.Pod); ok {
		for _, v := range pods {
			c.Total += 1
			switch corev1.PodPhase(v.Status.Phase) {
			case corev1.PodPending:
				c.Pending += 1
			case corev1.PodFailed:
				c.Failed += 1
			case corev1.PodSucceeded:
				c.Succeeded += 1
				for _, s := range v.Status.Conditions {
					if s.Type == corev1.PodReady && s.Status == corev1.ConditionTrue {
						c.Ready += 1
					}
				}
			case corev1.PodRunning:
				c.Running += 1
			case corev1.PodUnknown:
				c.Unknown += 1
			default:
				c.Unknown += 1
			}
		}
	}
	return c
}
