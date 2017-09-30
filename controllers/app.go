package controllers

type AppController struct {
	baseController
}

// newApps
// @Title 应用
// @Description   添加新的应用
// @Param Token header string true 'Token'
// @Param body body string true "应用配置"
// @Success 201 {string} create success!
// @Failure 500
// @router / [Post]
func (this *AppController) NewApp() {

	this.normalReturn("ok")
}
