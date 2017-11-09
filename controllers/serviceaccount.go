package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/serviceaccount"
	"ufleet-deploy/pkg/user"

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
		v, _ := pk.GetServiceAccountInterface(j)
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
		v, _ := pk.GetServiceAccountInterface(j)
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
		v, _ := pk.GetServiceAccountInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateServiceAccount
// @Title ServiceAccount
// @Description  创建服务帐号
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ServiceAccountController) CreateServiceAccount() {
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
	opt.Comment = co.Comment
	opt.User = who
	err = pk.Controller.CreateObject(group, workspace, []byte(co.Data), opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

type ServiceAccountCustomOption struct {
	Comment string   `json:"comment"`
	Name    string   `json:"name"`
	Secrets []string `json:"secrets"`
}

// CreateServiceAccountCustom
// @Title ServiceAccount
// @Description  创建服务帐号
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace/custom [Post]
func (this *ServiceAccountController) CreateServiceAccountCustom() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	var co ServiceAccountCustomOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

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
	opt.Comment = co.Comment
	opt.User = who
	err = pk.Controller.CreateObject(group, workspace, bytedata, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateServiceAccount
// @Title ServiceAccount
// @Description  更新服务帐号
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "服务帐号"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace [Put]
func (this *ServiceAccountController) UpdateServiceAccount() {
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
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, serviceaccount, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateServiceAccountCustom
// @Title ServiceAccount
// @Description  创建服务帐号
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace/custom [Put]
func (this *ServiceAccountController) UpdateServiceAccountCustom() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	sa := this.Ctx.Input.Param(":serviceaccount")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	var co ServiceAccountCustomOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	cm := corev1.ServiceAccount{}
	cm.Kind = "ServiceAccount"
	cm.APIVersion = "v1"
	cm.Name = sa
	cm.Secrets = make([]corev1.ObjectReference, 0)
	for _, v := range co.Secrets {
		var or corev1.ObjectReference
		or.Name = v
		cm.Secrets = append(cm.Secrets, or)
	}

	bytedata, err := json.Marshal(cm)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var opt resource.UpdateOption
	opt.Comment = co.Comment
	err = pk.Controller.UpdateObject(group, workspace, sa, bytedata, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteServiceAccount
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "服务帐号"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace [Delete]
func (this *ServiceAccountController) DeleteServiceAccount() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	err := pk.Controller.DeleteObject(group, workspace, serviceaccount, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
}

// GetServiceAccountTemplate
// @Title ServiceAccount
// @Description   ServiceAccount
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "服务帐号"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace/template [Get]
func (this *ServiceAccountController) GetServiceAccountTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	v, err := pk.Controller.GetObject(group, workspace, serviceaccount)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceAccountInterface(v)
	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetServiceAccountContainerEvents
// @Title ServiceAccount
// @Description   ServiceAccount container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "服务帐号"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace/event [Get]
func (this *ServiceAccountController) GetServiceAccountEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	v, err := pk.Controller.GetObject(group, workspace, endpoint)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceAccountInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}

// GetServiceAccountReferencObjects
// @Title ServiceAccount
// @Description   ServiceAccount container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param serviceaccount path string true "私密凭据"
// @Success 201 {string} create success!
// @Failure 500
// @router /:serviceaccount/group/:group/workspace/:workspace/reference [Get]
func (this *ServiceAccountController) GetServiceAccountReferenceObject() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	serviceaccount := this.Ctx.Input.Param(":serviceaccount")

	v, err := pk.Controller.GetObject(group, workspace, serviceaccount)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetServiceAccountInterface(v)
	es, err := pi.GetReferenceObjects()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
