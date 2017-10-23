package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/ingress"
)

type IngressController struct {
	baseController
}

// ListIngresss
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *IngressController) ListIngresss() {

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

// GetIngresss
// @Title Ingress
// @Description  Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "服务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace [Get]
func (this *IngressController) GetGroupWorkspaceIngress() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	pi, err := pk.Controller.Get(group, workspace, ingress)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi

	var js *pk.Status
	js = v.GetStatus()

	this.normalReturn(js)
}

// ListGroupsIngresss
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *IngressController) ListGroupsIngresss() {

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

	pis := make([]pk.IngressInterface, 0)

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

// ListGroupIngresss
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *IngressController) ListGroupIngresss() {

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

// CreateIngress
// @Title Ingress
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *IngressController) CreateIngress() {

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

	var opt resource.CreateOption
	err := pk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// UpdateIngress
// @Title Ingress
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace [Put]
func (this *IngressController) UpdateIngress() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, ingress, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteIngress
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace [Delete]
func (this *IngressController) DeleteIngress() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	err := pk.Controller.Delete(group, workspace, ingress, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetIngressContainerEvents
// @Title Ingress
// @Description   Ingress container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace/event [Get]
func (this *IngressController) GetIngressEvent() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	service := this.Ctx.Input.Param(":service")

	pi, err := pk.Controller.Get(group, workspace, service)
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
