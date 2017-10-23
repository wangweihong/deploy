package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/serviceaccount"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type ServiceAccountController struct {
	baseController
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
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
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

	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}
	this.normalReturn(jss)
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

	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
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

	var co models.CreateOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	var opt resource.CreateOption
	opt.Comment = co.Comment
	err = pk.Controller.Create(group, workspace, []byte(co.Data), opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

type ServiceAccountCreateOption struct {
	Comment string `json:"comment"`
	//	Data    string `json:"data"`
	//Data json.RawMessage `json:"data"`
	Name    string   `json:"name"`
	Secrets []string `json:"secrets"`
}

// CreateServiceAccountCustom
// @Title ServiceAccount
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace/custom [Post]
func (this *ServiceAccountController) CreateServiceAccountCustom() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	var co ServiceAccountCreateOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/
	cm := corev1.ServiceAccount{}
	cm.Kind = "ServiceAccount"
	cm.APIVersion = "v1"
	cm.Name = co.Name
	cm.Secrets = make([]corev1.ObjectReference, 0)
	for _, v := range co.Secrets {
		var or corev1.ObjectReference
		or.Name = v
		cm.Secrets = append(cm.Secrets, or)
	}

	bytedata, err := json.Marshal(cm)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	var opt resource.CreateOption
	opt.Comment = co.Comment
	err = pk.Controller.Create(group, workspace, bytedata, opt)
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

	err := pk.Controller.Delete(group, workspace, serviceaccount, resource.DeleteOption{})
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
