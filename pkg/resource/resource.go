package resource

import (
	"fmt"
	"sync"
)

var (
	resourceToRCUD = make(map[string]RCUD)
	locker         = sync.Mutex{}

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
	App  *string //所属app
	User string  //创建的用户
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
