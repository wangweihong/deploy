package controllers

import (
	"fmt"
	"ufleet-deploy/pkg/app"
)

type AppController struct {
	baseController
}

// newApps
// @Title 应用
// @Description   添加新的应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Post]
func (this *AppController) NewApp() {

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	var opt app.CreateOptions
	err := app.Controller.NewApp(group, workspace, appName, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn("ok")

}

// deleteApps
// @Title 应用
// @Description   删除应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Delete]
func (this *AppController) DeleteApp() {

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	err := app.Controller.DeleteApp(group, workspace, appName, app.DeleteOptions{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn("ok")
}

// GetApp
// @Title 应用
// @Description   添加指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Get]
func (this *AppController) GetApp() {

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	ai, err := app.Controller.Get(group, workspace, appName)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(ai.Info())

}

// ListGroupApp
// @Title 应用
// @Description   添加指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *AppController) ListGroupApp() {

	group := this.Ctx.Input.Param(":group")

	ais, err := app.Controller.List(group, app.ListOptions{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	infos := make([]app.App, 0)
	for _, v := range ais {
		infos = append(infos, v.Info())
	}

	this.normalReturn(infos)

}
