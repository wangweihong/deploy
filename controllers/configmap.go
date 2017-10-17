package controllers

import (
	"encoding/json"
	"fmt"
	sk "ufleet-deploy/pkg/resource/configmap"
	pk "ufleet-deploy/pkg/resource/statefulset"
)

type ConfigMapController struct {
	baseController
}

type ConfigMapState struct {
	Name       string `json:"name"`
	Group      string `json:"group"`
	Workspace  string `json:"workspace"`
	User       string `json:"user"`
	CreateTime int64  `json:"createtime"`
}

// ListConfigMaps
// @Title ConfigMap
// @Description  ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ConfigMapController) ListGroupWorkspaceConfigMaps() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := sk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	configmaps := make([]sk.ConfigMap, 0)
	for _, v := range pis {
		t := v.Info()
		configmaps = append(configmaps, *t)
	}

	this.normalReturn(configmaps)
}

// ListGroupsConfigMaps
// @Title ConfigMap
// @Description   ConfigMap
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ConfigMapController) ListGroupsConfigMaps() {

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

	pis := make([]sk.ConfigMapInterface, 0)

	for _, v := range groups {
		tmp, err := sk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]ConfigMapState, 0)
	for _, v := range pis {
		var ps ConfigMapState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User

	}

	this.normalReturn(pss)
}

// ListGroupConfigMaps
// @Title ConfigMap
// @Description   ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *ConfigMapController) ListGroupConfigMaps() {

	group := this.Ctx.Input.Param(":group")

	pis, err := sk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]ConfigMapState, 0)
	for _, v := range pis {
		var ps ConfigMapState
		ps.Name = v.Info().Name
		ps.Group = v.Info().Group
		ps.Workspace = v.Info().Workspace
		ps.User = v.Info().User

	}

	this.normalReturn(pss)
}

// CreateConfigMap
// @Title ConfigMap
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ConfigMapController) CreateConfigMap() {

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

	var opt sk.CreateOptions
	err := sk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteConfigMap
// @Title ConfigMap
// @Description   ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param configmap path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace [Delete]
func (this *ConfigMapController) DeleteConfigMap() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	err := sk.Controller.Delete(group, workspace, configmap, sk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// UpdateConfigMap
// @Title ConfigMap
// @Description  更新配置表
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param configmap path string true "配置表"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace [Put]
func (this *ConfigMapController) UpdateConfigMap() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, configmap, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetConfigMapTemplate
// @Title ConfigMap
// @Description   ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param configmap path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace/template [Get]
func (this *ConfigMapController) GetConfigMapTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	pi, err := sk.Controller.Get(group, workspace, configmap)
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
