package controllers

import (
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
