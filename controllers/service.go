package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/service"
	"ufleet-deploy/pkg/user"
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
func (this *ServiceController) ListGroupWorkspaceServices() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetServiceInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetServices
// @Title Service
// @Description  Service
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param service path string true "服务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:service/group/:group/workspace/:workspace [Get]
func (this *ServiceController) GetGroupWorkspaceService() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	v, err := pk.Controller.GetObject(group, workspace, service)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceInterface(v)

	var js *pk.Status
	js = pi.GetStatus()

	this.normalReturn(js)
}

// ListGroupsServices
// @Title Service
// @Description   Service
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ServiceController) ListGroupsServices() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

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

	pis := make([]resource.Object, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetServiceInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupServices
// @Title Service
// @Description   Service
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *ServiceController) ListGroupServices() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	pis, err := pk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetServiceInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateService
// @Title Service
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ServiceController) CreateService() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	ui := user.NewUserClient(token)
	who, err := ui.GetUserName()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var opt resource.CreateOption
	opt.User = who
	err = pk.Controller.CreateObject(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateService
// @Title Service
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param service path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:service/group/:group/workspace/:workspace [Put]
func (this *ServiceController) UpdateService() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		this.audit(token, "", true)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		this.audit(token, "", true)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, service, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		this.audit(token, "", true)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
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
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	err := pk.Controller.DeleteObject(group, workspace, service, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	v, err := pk.Controller.GetObject(group, workspace, service)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceInterface(v)

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetServiceContainerEvents
// @Title Service
// @Description   Service container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param service path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:service/group/:group/workspace/:workspace/event [Get]
func (this *ServiceController) GetServiceEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	v, err := pk.Controller.GetObject(group, workspace, service)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
