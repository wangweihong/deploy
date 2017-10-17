package controllers

import (
	"encoding/json"
	"fmt"
	pk "ufleet-deploy/pkg/resource/serviceaccount"
)

type ServiceAccountController struct {
	baseController
}

type ServiceAccountState struct {
	Name       string `json:"name"`
	Group      string `json:"group"`
	Workspace  string `json:"workspace"`
	User       string `json:"user"`
	CreateTime int64  `json:"createtime"`
}

// ListServiceAccounts
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ServiceAccountController) ListServiceAccounts() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	serviceaccounts := make([]pk.ServiceAccount, 0)
	for _, v := range pis {
		t := v.Info()
		serviceaccounts = append(serviceaccounts, *t)
	}

	this.normalReturn(serviceaccounts)
}

// ListGroupsServiceAccounts
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ServiceAccountController) ListGroupsServiceAccounts() {

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

	pis := make([]pk.ServiceAccountInterface, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]ServiceAccountState, 0)
	for _, v := range pis {
		var ps ServiceAccountState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User
	}

	this.normalReturn(pss)
}

// ListGroupServiceAccounts
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *ServiceAccountController) ListGroupServiceAccounts() {

	group := this.Ctx.Input.Param(":group")

	pis, err := pk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]ServiceAccountState, 0)
	for _, v := range pis {
		var ps ServiceAccountState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User
	}

	this.normalReturn(pss)
}

// CreateServiceAccount
// @Title ServiceAccount
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ServiceAccountController) CreateServiceAccount() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	var opt pk.CreateOptions
	err := pk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// UpdateServiceAccount
// @Title ServiceAccount
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace [Put]
func (this *ServiceAccountController) UpdateServiceAccount() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, serviceaccount, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteServiceAccount
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace [Delete]
func (this *ServiceAccountController) DeleteServiceAccount() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	err := pk.Controller.Delete(group, workspace, serviceaccount, pk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetServiceAccountTemplate
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace/template [Get]
func (this *ServiceAccountController) GetServiceAccountTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	pi, err := pk.Controller.Get(group, workspace, serviceaccount)
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
