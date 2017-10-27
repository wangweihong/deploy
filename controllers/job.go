package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/job"
	sk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/user"
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
func (this *JobController) ListGroupWorkspaceJobs() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//jobs := make([]pk.Job, 0)
	jss := make([]pk.Status, 0)

	for _, j := range pis {
		v, _ := pk.GetJobInterface(j)
		js, err := v.GetStatus()
		if err != nil {
			js := &pk.Status{}
			job := v.Info()
			js.Name = job.Name
			js.User = job.User
			js.Workspace = job.Workspace
			js.Group = job.Group
			js.Reason = err.Error()
			js.PodStatus = make([]sk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetJob
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace [Get]
func (this *JobController) GetJob() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	pi, err := pk.Controller.GetObject(group, workspace, job)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetJobInterface(pi)

	js, err := v.GetStatus()
	if err != nil {
		js = &pk.Status{}
		job := v.Info()
		js.Name = job.Name
		js.User = job.User
		js.Workspace = job.Workspace
		js.Group = job.Group
		js.Reason = err.Error()
		js.PodStatus = make([]sk.Status, 0)
	}

	this.normalReturn(js)
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

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

	pis := make([]resource.Object, 0)
	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//jobs := make([]pk.Job, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetJobInterface(j)
		js, err := v.GetStatus()
		if err != nil {
			js := &pk.Status{}
			job := v.Info()
			js.Name = job.Name
			js.User = job.User
			js.Workspace = job.Workspace
			js.Group = job.Group
			js.Reason = err.Error()
			js.PodStatus = make([]sk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupJobs
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *JobController) ListGroupJobs() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	pis, err := pk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//jobs := make([]pk.Job, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetJobInterface(j)
		js, err := v.GetStatus()
		if err != nil {
			js := &pk.Status{}
			job := v.Info()
			js.Name = job.Name
			js.User = job.User
			js.Workspace = job.Workspace
			js.Group = job.Group
			js.Reason = err.Error()
			js.PodStatus = make([]sk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateJob
// @Title Job
// @Description  创建一次性任务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *JobController) CreateJob() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		this.audit(token, "", true)
		err := fmt.Errorf("must commit resource json/yaml data")
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

	var opt resource.CreateOption
	opt.User = who

	err = pk.Controller.CreateObject(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateJob
// @Title Job
// @Description  更新Job
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "Job"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace [Put]
func (this *JobController) UpdateJob() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, job, this.Ctx.Input.RequestBody)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
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
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	err := pk.Controller.DeleteObject(group, workspace, job, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetJobTemplate
// @Title Job
// @Description Get Job Template
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "任务名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace/template [Get]
func (this *JobController) GetJobTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	j, err := pk.Controller.GetObject(group, workspace, job)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	ji, _ := pk.GetJobInterface(j)

	t, err := ji.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetJobEvent
// @Title Job
// @Description   Job container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "一次性任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace/event [Get]
func (this *JobController) GetJobEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	j, err := pk.Controller.GetObject(group, workspace, job)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetJobInterface(j)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
