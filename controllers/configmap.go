package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/configmap"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type ConfigMapController struct {
	baseController
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

	pis := make([]pk.ConfigMapInterface, 0)

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

// CreateConfigMap
// @Title ConfigMap
// @Description  创建配置
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ConfigMapController) CreateConfigMap() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}
	/*
		var co models.CreateOption
		err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
		if err != nil {
			this.errReturn(err, 500)
			return
		}

	*/
	var opt resource.CreateOption
	//	opt.Comment = co.Comment
	err := pk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

type ConfigMapCreateOption struct {
	Comment string `json:"comment"`
	//	Data    string `json:"data"`
	Data map[string]string `json:"data"`
	//Data json.RawMessage `json:"data"`
	Name string `json:"name"`
}

// CreateConfigMapV1
// @Title ConfigMap
// @Description  创建配置
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace/custom [Post]
func (this *ConfigMapController) CreateConfigMapV1() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}
	var co ConfigMapCreateOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	cm := corev1.ConfigMap{}
	cm.Name = co.Name
	cm.Data = co.Data
	cm.Kind = "ConfigMap"
	cm.APIVersion = "v1"

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

// DeleteConfigMap
// @Title ConfigMap
// @Description   ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param configmap path string true "配置"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace [Delete]
func (this *ConfigMapController) DeleteConfigMap() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	err := pk.Controller.Delete(group, workspace, configmap, resource.DeleteOption{})
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
// @Param configmap path string true "配置"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace/template [Get]
func (this *ConfigMapController) GetConfigMapTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	pi, err := pk.Controller.Get(group, workspace, configmap)
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
