package controllers

import (
	"fmt"
	"ufleet-deploy/pkg/app"
	"ufleet-deploy/pkg/user"
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
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	appName := this.Ctx.Input.Param(":app")
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
	var opt app.CreateOption
	opt.User = who

	err = app.Controller.NewApp(group, workspace, appName, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, appName, false)
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
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	err := app.Controller.DeleteApp(group, workspace, appName, app.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, appName, false)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")

	ais, err := app.Controller.List(group, app.ListOption{})
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
