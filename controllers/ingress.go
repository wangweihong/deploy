package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/ingress"
	"ufleet-deploy/pkg/user"
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListGroupWorkspaceObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetIngressInterface(j)
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	pi, err := pk.Controller.GetObject(group, workspace, ingress)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetIngressInterface(pi)

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
		tmp, err := pk.Controller.ListGroupObject(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetIngressInterface(j)

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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetIngressInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateIngress
// @Title Ingress
// @Description  创建路由
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *IngressController) CreateIngress() {
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

// UpdateIngress
// @Title Ingress
// @Description  更新路由
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "路由"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace [Put]
func (this *IngressController) UpdateIngress() {
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
	ingress := this.Ctx.Input.Param(":ingress")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, ingress, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteIngress
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "路由"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace [Delete]
func (this *IngressController) DeleteIngress() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	err := pk.Controller.DeleteObject(group, workspace, ingress, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetIngressContainerEvents
// @Title Ingress
// @Description   Ingress container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "路由"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace/event [Get]
func (this *IngressController) GetIngressEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	v, err := pk.Controller.GetObject(group, workspace, ingress)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetIngressInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}

// GetIngressTemplate
// @Title Ingress
// @Description   Ingress
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param ingress path string true "路由"
// @Success 201 {string} create success!
// @Failure 500
// @router /:ingress/group/:group/workspace/:workspace/template [Get]
func (this *IngressController) GetIngressTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	v, err := pk.Controller.GetObject(group, workspace, ingress)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetIngressInterface(v)
	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
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
// @router /:ingress/group/:group/workspace/:workspace/services [Get]
func (this *IngressController) GetGroupWorkspaceIngressServices() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	ingress := this.Ctx.Input.Param(":ingress")

	pi, err := pk.Controller.GetObject(group, workspace, ingress)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetIngressInterface(pi)
	ss, err := v.GetServices()
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn(ss)
}
