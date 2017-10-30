package app

import (
	"fmt"
	"strings"
)

var (
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
