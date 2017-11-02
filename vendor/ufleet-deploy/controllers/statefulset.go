package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/statefulset"
	"ufleet-deploy/pkg/user"
)

type StatefulSetController struct {
	baseController
}

// ListStatefulSets
// @Title StatefulSet
// @Description   StatefulSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *StatefulSetController) ListStatefulSets() {
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
		v, _ := pk.GetStatefulSetInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetStatefulSets
// @Title StatefulSet
// @Description  StatefulSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param statefulset path string true "服务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:statefulset/group/:group/workspace/:workspace [Get]
func (this *StatefulSetController) GetGroupWorkspaceStatefulSet() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	v, err := pk.Controller.GetObject(group, workspace, statefulset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetStatefulSetInterface(v)

	var js *pk.Status
	js = pi.GetStatus()

	this.normalReturn(js)
}

// ListGroupsStatefulSets
// @Title StatefulSet
// @Description   StatefulSet
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *StatefulSetController) ListGroupsStatefulSets() {
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
		v, _ := pk.GetStatefulSetInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupStatefulSets
// @Title StatefulSet
// @Description   StatefulSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *StatefulSetController) ListGroupStatefulSets() {
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
		v, _ := pk.GetStatefulSetInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateStatefulSet
// @Title StatefulSet
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *StatefulSetController) CreateStatefulSet() {
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
		this.errReturn(err, 500)
		this.audit(token, "", true)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateStatefulSet
// @Title StatefulSet
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param statefulset path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:statefulset/group/:group/workspace/:workspace [Put]
func (this *StatefulSetController) UpdateStatefulSet() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, statefulset, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteStatefulSet
// @Title StatefulSet
// @Description   StatefulSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param statefulset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:statefulset/group/:group/workspace/:workspace [Delete]
func (this *StatefulSetController) DeleteStatefulSet() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	err := pk.Controller.DeleteObject(group, workspace, statefulset, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetStatefulSetContainerEvents
// @Title StatefulSet
// @Description   StatefulSet container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param statefulset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:statefulset/group/:group/workspace/:workspace/event [Get]
func (this *StatefulSetController) GetStatefulSetEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	v, err := pk.Controller.GetObject(group, workspace, statefulset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetStatefulSetInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
