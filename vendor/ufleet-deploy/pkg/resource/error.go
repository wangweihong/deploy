package resource

import "fmt"

var (
	ErrResourceNotFound  = fmt.Errorf("resource not found")
	ErrResourceExists    = fmt.Errorf("resource has exist")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrWorkspaceExists   = fmt.Errorf("workspace has exist")
	ErrGroupNotFound     = fmt.Errorf("group not found")
	ErrGroupExists       = fmt.Errorf("group has exist")
)
