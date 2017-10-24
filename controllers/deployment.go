package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/deployment"
	"ufleet-deploy/pkg/user"
)

type DeploymentController struct {
	baseController
}

// ListDeployments
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *DeploymentController) ListGroupWorkspaceDeployments() {
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

	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupDeployments
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *DeploymentController) ListGroupDeployments() {
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
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)

}

// GetDeployments
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区名"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace [Get]
func (this *DeploymentController) GetDeployment() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi
	js := v.GetStatus()

	this.normalReturn(js)

}

// CreateDeployment
// @Title Deployment
// @Description  创建部署
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *DeploymentController) CreateDeployment() {
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

// UpdateDeployment
// @Title Deployment
// @Description  deploymnt
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "deployment"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace [Put]
func (this *DeploymentController) UpdateDeployment() {
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
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, deployment, this.Ctx.Input.RequestBody)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteDeployment
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace [Delete]
func (this *DeploymentController) DeleteDeployment() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	err := pk.Controller.Delete(group, workspace, deployment, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetDeploymentContainerEvents
// @Title Deployment
// @Description   Deployment container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/event [Get]
func (this *DeploymentController) GetDeploymentEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
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

// ScaleDeployment
// @Title Deployment
// @Description  扩容副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "副本控制器"
// @Param replicas path string true "副本数"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/replicas/:replicas [Put]
func (this *DeploymentController) ScaleDeployment() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	replicasStr := this.Ctx.Input.Param(":replicas")

	ri, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err = ri.Scale(int(replicas))
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// ScaleDeploymentIncremental
// @Title Deployment
// @Description  递增/减锁扩容
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "副本控制器"
// @Param increment path string true "增副本数"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/increment/:increment [Put]
func (this *DeploymentController) ScaleDeploymentIncrement() {
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
	deployment := this.Ctx.Input.Param(":deployment")
	incrementStr := this.Ctx.Input.Param(":increment")

	ri, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	increment, err := strconv.ParseInt(incrementStr, 10, 32)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	js := ri.GetStatus()
	if js.Reason != "" {
		err := fmt.Errorf(js.Reason)
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	newReplicas := js.Replicas + int32(increment)

	err = ri.Scale(int(newReplicas))
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetDeploymentTemplate
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/template [Get]
func (this *DeploymentController) GetDeploymentTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
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

type DeployReplicas struct {
	Revision   int    `json:"revision"`
	Name       string `json:"name"`
	Desire     int    `json:"desire"`
	Current    int    `json:"current"`
	Running    int    `json:"running"`
	CreateTime int64  `json:"createtime"`
}

// GetDeploymentReplicaset
// @Title Deployment
// @Description   Deployment 副本
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/replicasets [Get]
func (this *DeploymentController) GetDeploymentReplicaSet() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	rm, err := pi.GetAllReplicaSets()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	drs := make([]DeployReplicas, 0)
	for k, v := range rm {
		var dr DeployReplicas
		dr.Revision = int(k)
		dr.Name = v.Name
		if v.Spec.Replicas != nil {
			dr.Desire = int(*v.Spec.Replicas)
		} else {
			dr.Desire = 1
		}

		dr.Current = int(v.Status.Replicas)
		dr.Running = int(v.Status.AvailableReplicas)
		dr.CreateTime = v.CreationTimestamp.Unix()
		drs = append(drs, dr)

	}

	this.normalReturn(drs)
}

// GetDeploymentRevisionsAndDecribe
// @Title Deployment
// @Description   Deployment 版本描述
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/revisions [Get]
func (this *DeploymentController) GetDeploymentRevisions() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	rm, err := pi.GetRevisionsAndDescribe()
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	drs := make([]struct {
		Revision int    `json:"revision"`
		Describe string `json:"describe"`
	}, 0)

	for k, v := range rm {
		dr := struct {
			Revision int    `json:"revision"`
			Describe string `json:"describe"`
		}{}
		dr.Revision = int(k)
		dr.Describe = v
		drs = append(drs, dr)

	}

	this.normalReturn(drs)
}

// Rollback
// @Title Deployment
// @Description   Deployment回滚
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param revision path string true "版本"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/revision/:revision [Put]
func (this *DeploymentController) RollBack() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	revision := this.Ctx.Input.Param(":revision")

	toRevision, err := strconv.ParseInt(revision, 10, 64)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	result, err := pi.Rollback(toRevision)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return

	}

	this.audit(token, "", false)
	this.normalReturn(*result)
}

// GetHpa
// @Title Deployment
// @Description  获取 Deployment hpa
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/hpa [Get]
func (this *DeploymentController) GetHPA() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	result, err := pi.GetAutoScale()
	if err != nil {
		this.errReturn(err, 500)
		return

	}

	this.normalReturn(*result)
}

type HpaOptions struct {
	CpuPercent  int `json:"cpupercent"`
	MemPercent  int `json:"mempercent"`
	NetPercent  int `json:"netpercent"`
	DiskPercent int `json:"diskpercent"`
	MinReplicas int `json:"minreplicas"`
	MaxReplicas int `json:"Maxreplicas"`
}

// StartHpa
// @Title Deployment
// @Description  启动Deployment hpa
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param body body string true "弹性伸缩参数"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/hpa [Post]
func (this *DeploymentController) StartHPA() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit hpa options")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var hpaopt HpaOptions
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &hpaopt)
	if err != nil {
		err = fmt.Errorf("parse hpa option fail for  fail for ", err)
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err = pi.StartAutoScale(hpaopt.MinReplicas, hpaopt.MaxReplicas, hpaopt.CpuPercent, hpaopt.MemPercent, hpaopt.DiskPercent, hpaopt.NetPercent)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}
