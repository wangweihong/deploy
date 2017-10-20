package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	ck "ufleet-deploy/pkg/resource/cronjob"
	jk "ufleet-deploy/pkg/resource/job"
)

type CronJobController struct {
	baseController
}

type CronJobState struct {
	Name             string      `json:"name"`
	User             string      `json:"user"`
	Workspace        string      `json:"workspace"`
	Group            string      `json:"group"`
	Total            int         `json:"total"`
	Running          int         `json:"running"`
	LastScheduleTime int64       `json:"lastscheduletime"`
	period           string      `json:"period"`
	JobStatus        []jk.Status `json:"jobtatus"`
	//	Pods       []string `json:"pods"`
}

// ListCronJobs
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *CronJobController) ListGroupWorkspaceCronJobs() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := ck.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	cronjobs := make([]ck.CronJob, 0)
	for _, v := range pis {
		t := v.Info()
		cronjobs = append(cronjobs, *t)
	}

	this.normalReturn(cronjobs)
}

// GetCronJob
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace [Get]
func (this *CronJobController) GetCronJob() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	pi, err := ck.Controller.Get(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi
	var s *ck.Status
	s, err = v.GetStatus()
	if err != nil {
		info := v.Info()
		s = &ck.Status{}
		s.JobStatus = make([]jk.Status, 0)
		s.Name = info.Name
		s.Group = info.Group
		s.Workspace = info.Workspace
		s.User = info.User
		s.Reason = err.Error()

	}

	this.normalReturn(s)
}

// ListGroupsCronJobs
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// ListGroupsCronJobs
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *CronJobController) ListGroupsCronJobs() {

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

	pis := make([]ck.CronJobInterface, 0)

	for _, v := range groups {
		tmp, err := ck.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]ck.Status, 0)
	for _, v := range pis {
		var s *ck.Status
		var err error
		s, err = v.GetStatus()
		if err != nil {
			info := v.Info()
			s = &ck.Status{}
			s.JobStatus = make([]jk.Status, 0)
			s.Name = info.Name
			s.Group = info.Group
			s.Workspace = info.Workspace
			s.User = info.User
			s.Reason = err.Error()

			pss = append(pss, *s)
			continue
		}

		pss = append(pss, *s)
	}

	this.normalReturn(pss)
}

// ListGroupCronJobs
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *CronJobController) ListGroupCronJobs() {

	group := this.Ctx.Input.Param(":group")

	pis, err := ck.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]ck.Status, 0)
	for _, v := range pis {
		var s *ck.Status
		var err error
		s, err = v.GetStatus()
		if err != nil {
			info := v.Info()
			s = &ck.Status{}
			s.JobStatus = make([]jk.Status, 0)
			s.Name = info.Name
			s.Group = info.Group
			s.Workspace = info.Workspace
			s.User = info.User
			s.Reason = err.Error()

			pss = append(pss, *s)
			continue
		}

		pss = append(pss, *s)
	}

	this.normalReturn(pss)
}

// CreateCronJob
// @Title CronJob
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *CronJobController) CreateCronJob() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	var opt resource.CreateOption
	err := ck.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// UpdateCronJob
// @Title CronJob
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace [Put]
func (this *CronJobController) UpdateCronJob() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := ck.Controller.Update(group, workspace, cronjob, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteCronJob
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace [Delete]
func (this *CronJobController) DeleteCronJob() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	err := ck.Controller.Delete(group, workspace, cronjob, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetCronJobTemplate
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace/template [Get]
func (this *CronJobController) GetCronJobTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	pi, err := ck.Controller.Get(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}
