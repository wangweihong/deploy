package controllers

import (
	pk "ufleet-deploy/pkg/resource/endpoint"
)

type EndpointController struct {
	baseController
}

// ListEndpoints
// @Title Endpoint
// @Description   Endpoint
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *EndpointController) ListEndpoints() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	endpoints := make([]pk.Endpoint, 0)
	for _, v := range pis {
		t := v.Info()
		endpoints = append(endpoints, *t)
	}

	this.normalReturn(endpoints)
}
