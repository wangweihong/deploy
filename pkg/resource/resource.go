package resource

import (
	"fmt"
	"sync"
)

var (
	resourceToRCUD = make(map[string]RCUD)
	locker         = sync.Mutex{}

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
}

type RCUD interface {
	Create(group, workspace string, data []byte, opt CreateOption) error
	Delete(group, workspace, resource string, opt DeleteOption) error
	Update(group, workspace, resource string, newdata []byte) error
}

func RegisterCURInterface(name string, cud RCUD) error {
	locker.Lock()
	defer locker.Unlock()

	if _, ok := resourceToRCUD[name]; ok {
		return fmt.Errorf("resource %v has register to RCUD", name)
	}

	resourceToRCUD[name] = cud
	return nil

}

func GetResourceCUD(name string) (RCUD, error) {
	locker.Lock()
	defer locker.Unlock()

	cud, ok := resourceToRCUD[name]
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
	Delete(group, workspace, name string, opt DeleteOption) error
	GetObjectWithoutLock(group, workspace, name string) (Object, error)
	DeleteGroup(group string) error
	DeleteWorkspace(groupName string, workspaceName string) error
	AddGroup(group string) error
	AddWorkspace(group, workspace string) error
	AddObjectFromBytes(data []byte) error
}
