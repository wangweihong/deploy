package controllers

import (
	"encoding/json"
	"fmt"
	jk "ufleet-deploy/pkg/resource/job"
	pk "ufleet-deploy/pkg/resource/pod"
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
func (this *JobController) ListGroupJobs() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := jk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//jobs := make([]jk.Job, 0)
	jss := make([]jk.Status, 0)

	for _, v := range pis {
		js := &jk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			job := v.Info()
			js.Name = job.Name
			js.User = job.User
			js.Workspace = job.Workspace
			js.Group = job.Group
			js.Reason = err.Error()
			js.PodStatus = make([]pk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupsJobs
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *JobController) ListGroupsJobs() {

	groups := make([]string, 0)
	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.errReturn(err, 500)
		return
	}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &groups)
	if err != nil {
		err = fmt.Errorf("try to unmarshal data \"%v\" fail for %v", string(this.Ctx.Input.RequestBody), err)
		this.errReturn(err, 500)
		return
	}

	pis := make([]jk.JobInterface, 0)
	for _, v := range groups {
		tmp, err := jk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//jobs := make([]jk.Job, 0)
	jss := make([]jk.Status, 0)
	for _, v := range pis {
		js := &jk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			job := v.Info()
			js.Name = job.Name
			js.User = job.User
			js.Workspace = job.Workspace
			js.Group = job.Group
			js.Reason = err.Error()
			js.PodStatus = make([]pk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// DeleteJob
// @Title Job
// @Description  DeleteeJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "任务名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace [Delete]
func (this *JobController) DeleteJob() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	err := jk.Controller.Delete(group, workspace, job, jk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
