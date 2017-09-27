package app

import (
	"fmt"
	"strings"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/option"
	"ufleet-deploy/pkg/resource"
)

var (
	Controller AppController
	sm         *AppMananger

	ErrAppNotFound       = fmt.Errorf("app not found")
	ErrAppExists         = fmt.Errorf("app exists")
	ErrGroupNotFound     = fmt.Errorf("group not found")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
)

func IsAppNotFound(err error) bool {
	if strings.HasPrefix(err.Error(), ErrAppNotFound.Error()) {
		return true
	}
	return false
}

func IsAppExists(err error) bool {
	if strings.HasPrefix(err.Error(), ErrAppExists.Error()) {
		return true
	}
	return false
}

type AppGetter interface {
	AppIF(group, workspaceName, name string) (AppInterface, error)
}

type AppLister interface {
	AppIFs(group string, opt ListOptions) ([]AppInterface, error)
}

type AppInterface interface {
	GetTemplates()
	GetResources()
	AddResources()
	RemoveResource(kind string, name string, flush bool) error
	Info() App
}

func generateResourceKey(kind string, name string) string {
	return kind + "_" + name

}
func (s *App) GetTemplates() {
}

func (s *App) GetResources() {}
func (s *App) AddResources() {}

func (s *App) Info() App {
	return *s
}

//更新etcd的数据
func (s *App) RemoveResource(kind string, name string, flush bool) error {
	key := generateResourceKey(kind, name)
	_, ok := s.Resources[key]
	if !ok {
		return ErrAppNotFound
	}

	rc, err := resource.GetResourceControllerFromKind(kind)
	if err != nil {
		return err
	}

	opt := option.DeleteOption{}
	opt.Group = s.Group
	opt.Workspace = s.Name
	opt.Name = name
	err = rc.Delete(opt)
	//忽略not found
	if err != nil || !resource.IsErrorNotFound(err) {
		return log.DebugPrint(err)
	}
	delete(s.Resources, key)

	//flush - 刷新到后端存储
	if flush {
		err := storer.Update(s.Group, s.Workspace, s.Name, s)
		if err != nil {
			return log.DebugPrint(err)
		}
	}

	return nil
}

type AppGroup struct {
	Workspaces map[string]AppWorkspace `json:"workspaces"`
}

type AppWorkspace struct {
	Apps map[string]App `json:"apps"`
}

type App struct {
	Templates []string            `json:"templates"` //各个resource的模板集合,yaml
	Name      string              `json:"name"`
	Group     string              `json:"group"`
	Workspace string              `json:"workspace"`
	Resources map[string]Resource `json:"resources"` //key: resourceKind_name
}

type Resource struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}
