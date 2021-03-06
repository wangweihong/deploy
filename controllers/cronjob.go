package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/cronjob"
	"ufleet-deploy/pkg/user"
)

type CronJobController struct {
	baseController
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

	pis, err := pk.Controller.ListGroupWorkspaceObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetCronJobInterface(j)
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

	pi, err := pk.Controller.GetObject(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetCronJobInterface(pi)
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

	pis := make([]resource.Object, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroupObject(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetCronJobInterface(j)

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

	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetCronJobInterface(j)
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
	err = pk.Controller.CreateObject(group, workspace, this.Ctx.Input.RequestBody, opt)
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
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, cronjob, true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, cronjob, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, cronjob, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, cronjob, false)
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
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	err := pk.Controller.DeleteObject(group, workspace, cronjob, resource.DeleteOption{})
	if err != nil {
		this.audit(token, cronjob, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, cronjob, false)
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

	v, err := pk.Controller.GetObject(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetCronJobInterface(v)

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

	v, err := pk.Controller.GetObject(group, workspace, cronjob)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetCronJobInterface(v)
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
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	cronjob := this.Ctx.Input.Param(":cronjob")

	v, err := pk.Controller.GetObject(group, workspace, cronjob)
	if err != nil {
		this.audit(token, cronjob, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetCronJobInterface(v)
	err = pi.SuspendOrResume()
	if err != nil {
		this.audit(token, cronjob, true)
		this.errReturn(err, 500)
		return

	}

	this.audit(token, cronjob, false)
	this.normalReturn("ok")
}
