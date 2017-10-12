package controllers

import (
	"encoding/json"
	"fmt"
	pk "ufleet-deploy/pkg/resource/cronjob"
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

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	cronjobs := make([]pk.CronJob, 0)
	for _, v := range pis {
		t := v.Info()
		cronjobs = append(cronjobs, *t)
	}

	this.normalReturn(cronjobs)
}

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

	pis := make([]pk.CronJobInterface, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]pk.Status, 0)
	for _, v := range pis {
		var s *pk.Status
		var err error
		s, err = v.GetStatus()
		if err != nil {
			info := v.Info()
			s = &pk.Status{}
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

	pis, err := pk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]pk.Status, 0)
	for _, v := range pis {
		var s *pk.Status
		var err error
		s, err = v.GetStatus()
		if err != nil {
			info := v.Info()
			s = &pk.Status{}
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

	err := pk.Controller.Delete(group, workspace, cronjob, pk.DeleteOption{})
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

	pi, err := pk.Controller.Get(group, workspace, cronjob)
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
