package resource

import (
	"strings"
	"ufleet-deploy/pkg/option"
)

type Controller interface {
	Create(data string, opt option.CreateOption) error
	Delete(opt option.DeleteOption) error
}

func GetResourceControllerFromKind(kind string) (Controller, error) {
	return nil, nil
}

func IsErrorNotFound(err error) bool {
	keyword := "not found"
	return strings.Contains(err.Error(), keyword)
}
