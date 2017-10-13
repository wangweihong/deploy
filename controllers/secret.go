package controllers

import (
	"encoding/json"
	"fmt"
	sk "ufleet-deploy/pkg/resource/secret"
	pk "ufleet-deploy/pkg/resource/statefulset"
)

type SecretController struct {
	baseController
}

type SecretState struct {
	Name       string `json:"name"`
	Group      string `json:"group"`
	Workspace  string `json:"workspace"`
	User       string `json:"user"`
	CreateTime int64  `json:"createtime"`
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

	pis, err := sk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	secrets := make([]sk.Secret, 0)
	for _, v := range pis {
		t := v.Info()
		secrets = append(secrets, *t)
	}

	this.normalReturn(secrets)
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

	pis := make([]sk.SecretInterface, 0)

	for _, v := range groups {
		tmp, err := sk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]SecretState, 0)
	for _, v := range pis {
		var ps SecretState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User

	}

	this.normalReturn(pss)
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
	pis, err := sk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pss := make([]SecretState, 0)
	for _, v := range pis {
		var ps SecretState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User

	}

	this.normalReturn(pss)
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

	err := sk.Controller.Delete(group, workspace, secret, sk.DeleteOption{})
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

	pi, err := sk.Controller.Get(group, workspace, secret)
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
