package controllers

import (
	pk "ufleet-deploy/pkg/pod"
)

type PodController struct {
	baseController
}

// ListPods
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *PodController) ListPods() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pods := make([]pk.Pod, 0)
	for _, v := range pis {
		t := v.Info()
		pods = append(pods, *t)
	}

	this.normalReturn(pods)
}
