package controllers

import (
	pk "ufleet-deploy/pkg/resource/job"
)

type JobController struct {
	baseController
}

// ListJobs
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *JobController) ListJobs() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jobs := make([]pk.Job, 0)
	for _, v := range pis {
		t := v.Info()
		jobs = append(jobs, *t)
	}

	this.normalReturn(jobs)
}
