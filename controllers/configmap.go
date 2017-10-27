package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/configmap"
	"ufleet-deploy/pkg/user"

	yaml "gopkg.in/yaml.v2"

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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {

		v, _ := pk.GetConfigMapInterface(j)
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
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetConfigMapInterface(j)
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
	for _, j := range pis {
		v, _ := pk.GetConfigMapInterface(j)
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

type ConfigMapCustomOption struct {
	Comment string `json:"comment"`
	//Data map[string]string `json:"data"`
	Data string `json:"data"`
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
func (this *ConfigMapController) CreateConfigMapCustom() {
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
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	var co ConfigMapCustomOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	data := make(map[string]string)
	err = yaml.Unmarshal([]byte(co.Data), &data)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	cm := corev1.ConfigMap{}
	cm.Name = co.Name
	//	cm.Data = co.Data
	cm.Data = data
	cm.Kind = "ConfigMap"
	cm.APIVersion = "v1"

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

// UpdateConfigMapCustom
// @Title ConfigMap
// @Description  创建配置
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace/custom [Put]
func (this *ConfigMapController) UpdateConfigMapCustom() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	var co ConfigMapCustomOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	data := make(map[string]string)
	err = yaml.Unmarshal([]byte(co.Data), &data)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	cm := corev1.ConfigMap{}
	cm.Name = configmap
	//	cm.Data = co.Data
	cm.Data = data
	cm.Kind = "ConfigMap"
	cm.APIVersion = "v1"

	bytedata, err := json.Marshal(cm)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var opt resource.UpdateOption
	opt.Comment = co.Comment

	err = pk.Controller.UpdateObject(group, workspace, co.Name, bytedata, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
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
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	err := pk.Controller.DeleteObject(group, workspace, configmap, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, configmap, false)
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
	configmap := this.Ctx.Input.Param(":configmap")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, configmap, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, configmap, false)
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	ri, err := pk.Controller.GetObject(group, workspace, configmap)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetConfigMapInterface(ri)

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetConfigMapContainerEvents
// @Title ConfigMap
// @Description   ConfigMap container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param configmap path string true "配置"
// @Success 201 {string} create success!
// @Failure 500
// @router /:configmap/group/:group/workspace/:workspace/event [Get]
func (this *ConfigMapController) GetConfigMapEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	configmap := this.Ctx.Input.Param(":configmap")

	ri, err := pk.Controller.GetObject(group, workspace, configmap)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetConfigMapInterface(ri)

	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
