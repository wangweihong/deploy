package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/cronjob"
	sk "ufleet-deploy/pkg/resource/job"
	"ufleet-deploy/pkg/user"
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
	JobStatus        []sk.Status `json:"jobtatus"`
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]pk.Status, 0)
	for _, v := range pis {
		s := v.GetStatus()
		pss = append(pss, *s)
	}

	this.normalReturn(pss)
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
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	pi, err := pk.Controller.Get(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi
	s := v.GetStatus()

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
		s := v.GetStatus()

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
	pss := make([]pk.Status, 0)
	for _, v := range pis {
		s := v.GetStatus()
		pss = append(pss, *s)
	}

	this.normalReturn(pss)
}

// CreateCronJob
// @Title CronJob
// @Description  创建定时任务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *CronJobController) CreateCronJob() {
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
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
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
	err = pk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateCronJob
// @Title CronJob
// @Description  更新定时任务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "定时任务"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace [Put]
func (this *CronJobController) UpdateCronJob() {
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
	cronjob := this.Ctx.Input.Param(":cronjob")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, cronjob, this.Ctx.Input.RequestBody)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteCronJob
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "定时任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace [Delete]
func (this *CronJobController) DeleteCronJob() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	err := pk.Controller.Delete(group, workspace, cronjob, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetCronJobTemplate
// @Title CronJob
// @Description   CronJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "定时任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace/template [Get]
func (this *CronJobController) GetCronJobTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

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

// GetCronJobEvent
// @Title CronJob
// @Description   CronJob container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "定时任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace/event [Get]
func (this *CronJobController) GetCronJobEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	pi, err := pk.Controller.Get(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}

// SuspendCronJob
// @Title CronJob
// @Description  暂停定时任务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param cronjob path string true "定时任务"
// @Success 201 {string} create success!
// @Failure 500
// @router /:cronjob/group/:group/workspace/:workspace/suspendandresume [Put]
func (this *CronJobController) SuspendOrResumeCronJob() {
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
	cronjob := this.Ctx.Input.Param(":cronjob")

	pi, err := pk.Controller.Get(group, workspace, cronjob)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	err = pi.SuspendOrResume()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return

	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}
