package controllers

import (
	"encoding/json"
	"fmt"
	sk "ufleet-deploy/pkg/resource/service"
)

type ServiceController struct {
	baseController
}

type ServiceState struct {
	Name       string `json:"name"`
	Group      string `json:"group"`
	Workspace  string `json:"workspace"`
	User       string `json:"user"`
	CreateTime int64  `json:"createtime"`
}

// ListServices
// @Title Service
// @Description  Service
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ServiceController) ListServices() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := sk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	services := make([]sk.Service, 0)
	for _, v := range pis {
		t := v.Info()
		services = append(services, *t)
	}

	this.normalReturn(services)
}

// ListGroupsServices
// @Title Service
// @Description   Service
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ServiceController) ListGroupServices() {

	groups := make([]string, 0)
	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.errReturn(err, 500)
		return
	}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &groups)
	if err != nil {
		err = fmt.Errorf("try to unmarshal data \"%v\" fail for %v", string(this.Ctx.Input.RequestBody), err)
		this.errReturn(err, 500)
		return
	}

	pis := make([]sk.ServiceInterface, 0)

	for _, v := range groups {
		tmp, err := sk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]ServiceState, 0)
	for _, v := range pis {
		var ps ServiceState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User

	}

	this.normalReturn(pss)
}

// DeleteService
// @Title Service
// @Description   Service
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param service path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:service/group/:group/workspace/:workspace [Delete]
func (this *ServiceController) DeleteService() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	err := sk.Controller.Delete(group, workspace, service, sk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetServiceTemplate
// @Title Service
// @Description   Service
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param service path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:service/group/:group/workspace/:workspace/template [Get]
func (this *ServiceController) GetServiceTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	pi, err := sk.Controller.Get(group, workspace, service)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}
