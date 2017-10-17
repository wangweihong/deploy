package controllers

import (
	"fmt"
	pk "ufleet-deploy/pkg/resource/statefulset"
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	statefulsets := make([]pk.StatefulSet, 0)
	for _, v := range pis {
		t := v.Info()
		statefulsets = append(statefulsets, *t)
	}

	this.normalReturn(statefulsets)
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

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, statefulset, this.Ctx.Input.RequestBody)
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	statefulset := this.Ctx.Input.Param(":statefulset")

	err := pk.Controller.Delete(group, workspace, statefulset, pk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
