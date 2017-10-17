package controllers

import (
	"fmt"
	pk "ufleet-deploy/pkg/resource/daemonset"
)

type DaemonSetController struct {
	baseController
}

// ListDaemonSets
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *DaemonSetController) ListDaemonSets() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	daemonsets := make([]pk.DaemonSet, 0)
	for _, v := range pis {
		t := v.Info()
		daemonsets = append(daemonsets, *t)
	}

	this.normalReturn(daemonsets)
}

// CreateDaemonSet
// @Title DaemonSet
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *DaemonSetController) CreateDaemonSet() {

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

// UpdateDaemonSet
// @Title DaemonSet
// @Description  更新daemonset组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "daemonset组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace [Put]
func (this *DaemonSetController) UpdateDaemonSet() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, daemonset, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteDaemonSet
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace [Delete]
func (this *DaemonSetController) DeleteDaemonSet() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	err := pk.Controller.Delete(group, workspace, daemonset, pk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
