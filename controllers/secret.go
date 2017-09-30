package controllers

import (
	sk "ufleet-deploy/pkg/resource/secret"
)

type SecretController struct {
	baseController
}

// ListSecrets
// @Title Secret
// @Description  Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *SecretController) ListSecrets() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := sk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	secrets := make([]sk.Secret, 0)
	for _, v := range pis {
		t := v.Info()
		secrets = append(secrets, *t)
	}

	this.normalReturn(secrets)
}
