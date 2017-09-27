package app

import (
	"encoding/json"
	"ufleet-deploy/deploy/backend"
)

const (
	backendKind = backend.ResourceApps
)

var (
	storer Store
)

type Store interface {
	Get(groupName, workspaceName, resouceName string) (*App, error)
	Create(groupName, workspaceName, resouceName string, data interface{}) error
	Update(groupName, workspaceName, resourceName string, data interface{}) error
	Delete(groupName, workspaceName, resourceName string) error
}

type store struct {
	backend.BackendHandler
}

func (s store) Get(groupName, workspaceName, resouceName string) (*App, error) {
	data, err := s.BackendHandler.GetResource(backendKind, groupName, workspaceName, resouceName)
	if err != nil {
		return nil, err
	}

	var a App
	err = json.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil

}

func (s store) Create(groupName, workspaceName, resouceName string, data interface{}) error {

	err := s.BackendHandler.CreateResource(backendKind, groupName, workspaceName, resouceName, data)
	if err != nil {
		return err
	}
	return nil
}

func (s store) Update(groupName, workspaceName, resouceName string, data interface{}) error {

	err := s.BackendHandler.UpdateResource(backendKind, groupName, workspaceName, resouceName, data)
	if err != nil {
		return err
	}
	return nil
}

func (s store) Delete(groupName, workspaceName, resourceName string) error {
	err := s.BackendHandler.DeleteResource(backendKind, groupName, workspaceName, resourceName)
	if err != nil {
		return err
	}
	return nil
}

func InitStore() Store {
	be := backend.NewBackendHandler()
	return store{BackendHandler: be}
}
