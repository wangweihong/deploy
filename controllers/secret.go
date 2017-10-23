package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/secret"
)

type SecretController struct {
	baseController
}

// ListSecrets
// @Title Secret
// @Description  Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *SecretController) ListSecrets() {

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

// ListGroupsSecrets
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *SecretController) ListGroupsSecrets() {

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

	pis := make([]pk.SecretInterface, 0)

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

// ListGroupSecrets
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *SecretController) ListGroupSecrets() {

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

// CreateSecret
// @Title Secret
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *SecretController) CreateSecret() {

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

// UpdateSecret
// @Title Secret
// @Description  更新secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "secret"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace [Put]
func (this *SecretController) UpdateSecret() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, secret, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteSecret
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace [Delete]
func (this *SecretController) DeleteSecret() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	err := pk.Controller.Delete(group, workspace, secret, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetSecretTemplate
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace/template [Get]
func (this *SecretController) GetSecretTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	pi, err := pk.Controller.Get(group, workspace, secret)
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
