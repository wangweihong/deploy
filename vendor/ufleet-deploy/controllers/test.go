package controllers

import (
	"ufleet-deploy/mode"
)

type TestController struct {
	baseController
}

// CheckDevMode
// @Title 模式
// @Description  查看当前是否处于开发模式
// @Success 201 {bool}
// @Failure 500
// @router /devmode [Get]
func (this *TestController) CheckDevMode() {
	if mode.IsDevelopMode() {
		this.normalReturn(true)
	} else {
		this.normalReturn(false)
	}
}
