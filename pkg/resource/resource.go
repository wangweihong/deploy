package resource

import (
	"fmt"
	"sync"
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
	Name       string `json:"name"`
	Workspace  string `json:"workspace"`
	Group      string `json:"group"`
	App        string `json:"app"`
	User       string `json:"user"`
	Template   string `json:"template"`
	CreateTime int64  `json:"createtime"`
	Comment    string `json:"comment"`
	MemoryOnly bool   `json:"memoryonly"`
}

type Object interface {
	Metadata() ObjectMeta
}
type Locker interface {
	Lock()
	Unlock()
}

type ObjectController interface {
	Locker
	NewObject(meta ObjectMeta) error
	GetObjectWithoutLock(group, workspace, name string) (Object, error)
	DeleteGroup(group string) error
	DeleteWorkspace(groupName string, workspaceName string) error
	AddGroup(group string) error
	AddWorkspace(group, workspace string) error
	AddObjectFromBytes(data []byte, force bool) error

	CreateObject(group, workspace string, data []byte, opt CreateOption) error
	DeleteObject(group, workspace, configmap string, opt DeleteOption) error
	GetObject(group, workspace, configmap string) (Object, error)
	UpdateObject(group, workspace, resource string, newdata []byte, opt UpdateOption) error
	ListObject(group, workspace string) ([]Object, error)
	ListGroup(group string) ([]Object, error)
}
