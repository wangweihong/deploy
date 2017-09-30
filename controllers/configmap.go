package controllers

import (
	sk "ufleet-deploy/pkg/resource/configmap"
)

type ConfigMapController struct {
	baseController
}

// ListConfigMaps
// @Title ConfigMap
// @Description  ConfigMap
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ConfigMapController) ListConfigMaps() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := sk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	configmaps := make([]sk.ConfigMap, 0)
	for _, v := range pis {
		t := v.Info()
		configmaps = append(configmaps, *t)
	}

	this.normalReturn(configmaps)
}
