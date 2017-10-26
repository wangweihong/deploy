package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	"ufleet-deploy/pkg/resource/util"
)

var (
	Controller AppController
	sm         *AppMananger

	ErrResourceNotFound  = fmt.Errorf("app not found")
	ErrResourceExists    = fmt.Errorf("app exists")
	ErrGroupNotFound     = fmt.Errorf("group not found")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
)

func IsAppNotFound(err error) bool {
	if strings.HasPrefix(err.Error(), ErrResourceNotFound.Error()) {
		return true
	}
	return false
}

func IsAppExists(err error) bool {
	if strings.HasPrefix(err.Error(), ErrResourceExists.Error()) {
		return true
	}
	return false
}

type AppInterface interface {
	GetTemplates()
	GetResources()
	AddResources([]byte, bool) error
	RemoveResource(kind string, name string, flush bool) error
	Info() App
}

func generateResourceKey(kind string, name string) string {
	return kind + "_" + name

}
func (s *App) GetTemplates() {
}

func (s *App) GetResources() {}
func (s *App) AddResources(desc []byte, flush bool) error {
	appName := s.Name
	groupName := s.Group
	workspaceName := s.Workspace
	be := backend.NewBackendHandler()
	exts, uerr := util.ParseJsonOrYaml(desc)
	if uerr != nil {
		return log.DebugPrint(uerr)
	}
	if len(exts) == 0 {
		return log.DebugPrint("must  offer  resource json/yaml data")
	}

	var err error
	for _, v := range exts {
		tmp := struct {
			Kind     string `json:"kind"`
			MetaData struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}{}
		var res Resource

		err = json.Unmarshal(v.Raw, &tmp)
		if err != nil {
			err = log.ErrorPrint("create app "+appName+" fail for %v", err)
			return err
		} else {
			if strings.TrimSpace(tmp.Kind) == "" || strings.TrimSpace(tmp.MetaData.Name) == "" {
				err = log.ErrorPrint("create app " + appName + " fail for resource kind or name not set")
				return err
			}

			res.Name = tmp.MetaData.Name
			res.Kind = tmp.Kind
		}
		key := generateResourceKey("Pod", res.Name)

		if _, ok := s.Resources[key]; ok {
			err = log.ErrorPrint("duplicate resource")
			return err
		}

		var rcud resource.RCUD
		rcud, err = resource.GetResourceCUD(res.Kind)
		if err != nil {
			return log.DebugPrint(err)
		}

		opt := resource.CreateOption{}
		opt.App = &appName
		err = rcud.Create(groupName, workspaceName, v.Raw, opt)
		if err != nil {
			return log.DebugPrint(err)
		}
		s.Resources[key] = res

		if flush {
			err = be.UpdateResource(backendKind, groupName, workspaceName, appName, s)
			if err != nil {
				err2 := rcud.Delete(groupName, workspaceName, res.Name, resource.DeleteOption{})
				if err2 != nil {
					log.ErrorPrint(err2)
				}
				return log.DebugPrint(err)
			}
		}

	}
	return nil
}

func (s *App) Info() App {
	return *s
}

//更新etcd的数据
//如果flush==true,则说明是一个删除应用触发的删除资源操作,
//不需要刷新到etcd,触发eventWatcher
func (s *App) RemoveResource(kind string, name string, flush bool) error {
	key := generateResourceKey(kind, name)
	_, ok := s.Resources[key]
	if !ok {
		return ErrResourceNotFound
	}

	rcud, err := resource.GetResourceCUD(kind)
	if err != nil {
		return err
	}

	opt := resource.DeleteOption{}
	if !flush {
		opt.DontCallApp = true
	}
	err = rcud.Delete(s.Group, s.Workspace, name, opt)
	if err != nil && !resource.IsErrorNotFound(err) {
		return err
	}
	if flush {
		delete(s.Resources, key)

		be := backend.NewBackendHandler()
		//		err := storer.Update(s.Group, s.Workspace, s.Name, s)
		err := be.UpdateResource(backendKind, s.Group, s.Workspace, s.Name, s)
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
	Name       string              `json:"name"`
	Group      string              `json:"group"`
	User       string              `json:"user"`
	Workspace  string              `json:"workspace"`
	CreateTime int64               `json:"createtime"`
	Resources  map[string]Resource `json:"resources"` //key: resourceKind_name
}

type Resource struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}
