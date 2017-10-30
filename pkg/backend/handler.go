package backend

import (
	"fmt"
	"path/filepath"

	"ufleet-deploy/pkg/kv"
	"ufleet-deploy/pkg/log"
)

var (
	BackendResourceNotFound      = fmt.Errorf("not found")
	BackendResourceAlreadyExists = fmt.Errorf("already exists")
	BackendResourceInvalid       = fmt.Errorf("invalid kind")
)

type ResourceGroup struct {
	Workspaces map[string]ResourceWorkspace
}

type ResourceWorkspace struct {
	Resources map[string]Resource
}

type Resource []byte //value

type BackendHandler interface {
	GetResource(kind, groupName, workspaceName, resouceName string) ([]byte, error)
	CreateResource(kind, groupName, workspaceName, resouceName string, data interface{}) error
	UpdateResource(kind, groupName, workspaceName, resourceName string, data interface{}) error
	DeleteResource(kind, groupName, workspaceName, resourceName string) error
	CreateResourceGroup(kind, groupName string) error
	DeleteResourceGroup(kind, groupName string) error
	CreateResourceWorkspace(kind, groupName, workspace string) error
	DeleteResourceWorkspace(kind, groupName, workspace string) error
	GetResourceAllGroup(kind string) (map[string]ResourceGroup, error)
	GetResourceGroupList(kind string) (map[string]string, error)
}

func NewBackendHandler() BackendHandler {
	return &eb{}
}

func generateBackendKey(kind string, opts ...string) (string, error) {
	resourceKey, ok := resourceToBackendkey[kind]
	if !ok {
		return "", fmt.Errorf("invalid kind %v", kind)
	}

	for _, v := range opts {
		resourceKey = resourceKey + "/" + v
	}
	return resourceKey, nil
}

type eb struct{}

func (e *eb) GetResource(kind, groupName, workspaceName, resouceName string) ([]byte, error) {
	key, err := generateBackendKey(kind, groupName, workspaceName, resouceName)
	if err != nil {
		return nil, BackendResourceInvalid
	}

	resp, err := kv.Store.GetNode(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return nil, BackendResourceNotFound
		}
	}
	return []byte(resp.Node.Value), nil

}

func (e *eb) CreateResource(kind, groupName, workspaceName, resouceName string, data interface{}) error {
	key, err := generateBackendKey(kind, groupName, workspaceName, resouceName)
	if err != nil {
		return BackendResourceInvalid
	}
	log.DebugPrint(key)

	_, err = kv.Store.CreateNode(key, data)
	if err != nil {
		if err == kv.ErrKeyAlreadyExists {
			return BackendResourceAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) UpdateResource(kind, groupName, workspaceName, resouceName string, data interface{}) error {
	key, err := generateBackendKey(kind, groupName, workspaceName, resouceName)
	if err != nil {
		return BackendResourceInvalid
	}

	_, err = kv.Store.UpdateNode(key, data)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return BackendResourceNotFound
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) DeleteResource(kind, groupName, workspaceName, resouceName string) error {
	key, err := generateBackendKey(kind, groupName, workspaceName, resouceName)
	if err != nil {
		return BackendResourceInvalid
	}

	_, err = kv.Store.DeleteNode(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return BackendResourceNotFound
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) CreateResourceGroup(kind, groupName string) error {
	key, err := generateBackendKey(kind, groupName)
	if err != nil {
		return BackendResourceInvalid
	}

	_, err = kv.Store.CreateDirNode(key)
	if err != nil {
		if err == kv.ErrKeyAlreadyExists {
			return BackendResourceAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) DeleteResourceGroup(kind, groupName string) error {
	key, err := generateBackendKey(kind, groupName)
	if err != nil {
		return BackendResourceInvalid
	}

	_, err = kv.Store.DeleteDirNode(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return BackendResourceNotFound
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) CreateResourceWorkspace(kind, groupName, workspace string) error {
	key, err := generateBackendKey(kind, groupName, workspace)
	if err != nil {
		return BackendResourceInvalid
	}

	//log.DebugPrint(key)
	_, err = kv.Store.CreateDirNode(key)
	if err != nil {
		if err == kv.ErrKeyAlreadyExists {
			return BackendResourceAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) DeleteResourceWorkspace(kind, groupName, workspace string) error {
	key, err := generateBackendKey(kind, groupName, workspace)
	if err != nil {
		return BackendResourceInvalid
	}

	//log.DebugPrint(key)
	_, err = kv.Store.DeleteDirNode(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return BackendResourceNotFound
		} else {
			return err
		}
	}

	return nil
}

func (e *eb) GetResourceAllGroup(kind string) (map[string]ResourceGroup, error) {
	key, err := generateBackendKey(kind)
	if err != nil {
		return nil, BackendResourceInvalid
	}

	resp, err := kv.Store.GetNode(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return nil, BackendResourceNotFound
		}
		return nil, err
	}

	groups := make(map[string]ResourceGroup)
	for _, v := range resp.Node.Nodes {
		groupName := filepath.Base(v.Key)

		var group ResourceGroup
		group.Workspaces = make(map[string]ResourceWorkspace)
		for _, j := range v.Nodes {
			workspaceName := filepath.Base(j.Key)
			var workspace ResourceWorkspace
			workspace.Resources = make(map[string]Resource)
			for _, n := range j.Nodes {
				resouceName := filepath.Base(n.Key)
				resource := []byte(n.Value)

				workspace.Resources[resouceName] = resource
			}
			group.Workspaces[workspaceName] = workspace

		}
		groups[groupName] = group

	}

	return groups, nil

}

func (e *eb) GetResourceGroupList(kind string) (map[string]string, error) {
	groups := make(map[string]string, 0)

	rs, err := e.GetResourceAllGroup(kind)
	if err != nil {
		if err == BackendResourceNotFound {
			return groups, nil
		}
		return nil, err
	}

	for k, _ := range rs {
		groups[k] = k
	}
	return groups, nil
}
