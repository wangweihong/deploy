package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/endpoint"
	"ufleet-deploy/pkg/user"
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetEndpoints
// @Title Endpoint
// @Description  Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "服务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace [Get]
func (this *EndpointController) GetGroupWorkspaceEndpoint() {

	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	pi, err := pk.Controller.Get(group, workspace, endpoint)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi

	var js *pk.Status
	js = v.GetStatus()

	this.normalReturn(js)
}

// ListGroupsEndpoints
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *EndpointController) ListGroupsEndpoints() {
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

	pis := make([]pk.EndpointInterface, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}

	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupEndpoints
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *EndpointController) ListGroupEndpoints() {
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
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateEndpoint
// @Title Endpoint
// @Description  创建端点
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *EndpointController) CreateEndpoint() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
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

	var co models.CreateOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.errReturn(err, 500)
		this.audit(token, "", true)
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
	opt.Comment = co.Comment
	opt.User = who
	err = pk.Controller.Create(group, workspace, []byte(co.Data), opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateEndpoint
// @Title Endpoint
// @Description  更新端点
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "端点"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace [Put]
func (this *EndpointController) UpdateEndpoint() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.Update(group, workspace, endpoint, this.Ctx.Input.RequestBody)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteEndpoint
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "端点"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace [Delete]
func (this *EndpointController) DeleteEndpoint() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	err := pk.Controller.Delete(group, workspace, endpoint, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetEndpointContainerEvents
// @Title Endpoint
// @Description   Endpoint container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param endpoint path string true "端点"
// @Success 201 {string} create success!
// @Failure 500
// @router /:endpoint/group/:group/workspace/:workspace/event [Get]
func (this *EndpointController) GetEndpointEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	pi, err := pk.Controller.Get(group, workspace, endpoint)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
