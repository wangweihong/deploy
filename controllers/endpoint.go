package controllers

import (
	"fmt"
	pk "ufleet-deploy/pkg/resource/endpoint"
)

type EndpointController struct {
	baseController
}

// ListEndpoints
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *EndpointController) ListEndpoints() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	endpoints := make([]pk.Endpoint, 0)
	for _, v := range pis {
		t := v.Info()
		endpoints = append(endpoints, *t)
	}

	this.normalReturn(endpoints)
}

// CreateEndpoint
// @Title Endpoint
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *EndpointController) CreateEndpoint() {

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

// UpdateEndpoint
// @Title Endpoint
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace [Put]
func (this *EndpointController) UpdateEndpoint() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, endpoint, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteEndpoint
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace [Delete]
func (this *EndpointController) DeleteEndpoint() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	err := pk.Controller.Delete(group, workspace, endpoint, pk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
