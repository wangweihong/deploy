package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/deployment"
	"ufleet-deploy/pkg/user"

	corev1 "k8s.io/api/core/v1"
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

	pis, err := pk.Controller.ListGroupWorkspaceObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetDeploymentInterface(j)
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

	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetDeploymentInterface(j)

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

	pi, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetDeploymentInterface(pi)
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
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, deployment, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
	this.normalReturn("ok")
}

// UpdateDeployment
// @Title Deployment
// @Description  deploymnt
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param body body string true "容器镜像"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/custom [Put]
func (this *DeploymentController) UpdateDeploymentCustom() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit container&image info")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	cis := make([]models.ContainerImage, 0)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &cis)
	if err != nil {
		err = fmt.Errorf("parse container&image fail for  fail for %v", err)
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	d := runtime.Deployment
	for _, j := range cis {
		var found bool
		for k, v := range d.Spec.Template.Spec.Containers {
			if v.Name == j.Container {
				d.Spec.Template.Spec.Containers[k].Image = j.Image
				found = true
				break
			}
		}
		if !found {
			err := fmt.Errorf("container '%v' not found in deployment '%v'", j.Container, deployment)
			this.audit(token, deployment, true)
			this.errReturn(err, 500)
			return
		}
	}

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
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

	err := pk.Controller.DeleteObject(group, workspace, deployment, resource.DeleteOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
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

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)
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

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	ri, _ := pk.GetDeploymentInterface(v)

	err = ri.Scale(int(replicas))
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
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
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	incrementStr := this.Ctx.Input.Param(":increment")

	increment, err := strconv.ParseInt(incrementStr, 10, 32)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	ri, _ := pk.GetDeploymentInterface(v)
	js := ri.GetStatus()
	if js.Reason != "" {
		err := fmt.Errorf(js.Reason)
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	newReplicas := js.Desire + int(increment)

	err = ri.Scale(int(newReplicas))
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
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

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetDeploymentInterface(v)

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

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetDeploymentInterface(v)

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

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetDeploymentInterface(v)

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
func (this *DeploymentController) RollBackDeployment() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	revision := this.Ctx.Input.Param(":revision")

	toRevision, err := strconv.ParseInt(revision, 10, 64)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	result, err := pi.Rollback(toRevision)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return

	}

	this.audit(token, deployment, false)
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

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	result, err := pi.GetAutoScale()
	if err != nil {
		this.errReturn(err, 500)
		return

	}

	this.normalReturn(*result)
}

// GetAllHpa
// @Title Deployment
// @Description  获取 all deployed Deployment hpa
// @Param Token header string true 'Token'
// @Success 201 {string} create success!
// @Failure 500
// @router /allgroup/hpas [Get]
func (this *DeploymentController) GetAllHPA() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}
	wdhpa := make([]struct {
		Group     string `json:"group"`
		Workspace string `json:"workspace"`
		Name      string `json:"name"`
		HPA       pk.HPA `json:"hpa"`
	}, 0)

	for _, group := range pk.Controller.ListGroups() {
		wd := struct {
			Group     string `json:"group"`
			Workspace string `json:"workspace"`
			Name      string `json:"name"`
			HPA       pk.HPA `json:"hpa"`
		}{}

		rs, err := pk.Controller.ListGroupObject(group)
		if err != nil {
			this.errReturn(err, 500)
			return
		}

		for _, v := range rs {
			pi, _ := pk.GetDeploymentInterface(v)
			s := pi.GetStatus()
			if s.Reason != "" {
				err := fmt.Errorf(s.Reason, 500)
				this.errReturn(err, 500)
				return
			}

			result, err := pi.GetAutoScale()
			if err != nil {
				this.errReturn(err, 500)
				return
			}

			if result.Deployed {
				wd.Group = group
				wd.Workspace = s.Workspace
				wd.Name = s.Name
				wd.HPA = *result

				wdhpa = append(wdhpa, wd)
			}
		}
	}

	this.normalReturn(wdhpa)
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
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit hpa options")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	var opt pk.HPA

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &opt)
	if err != nil {
		err = fmt.Errorf("parse hpa option fail for  fail for ", err)
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)
	err = pi.StartAutoScale(opt)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
	this.normalReturn("ok")
}

// Rollback Pause Or Resume
// @Title Deployment
// @Description   Deployment回滚暂停/恢复
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/resumeorpause [Put]
func (this *DeploymentController) RollBackResumeOrPauseDeployment() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	err = pi.ResumeOrPauseRollOut()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
	this.normalReturn("ok")
}

// GetDeploymentContainerEnv
// @Title Deployment
// @Description   Deployment Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/container/:container/env [Get]
func (this *DeploymentController) GetDeploymentContainerSpecEnv() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	stat := pi.GetStatus()
	if stat.Reason != "" {
		err := fmt.Errorf(stat.Reason)
		this.errReturn(err, 500)
		return
	}

	for _, v := range stat.ContainerSpecs {
		if v.Name == container {
			this.normalReturn(v.Env)
			return
		}
	}

	err = fmt.Errorf("container not found")

	this.errReturn(err, 500)
}

// DeploymentContainerEnv
// @Title Deployment
// @Description   新增Deployment Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/container/:container/env [Post]
func (this *DeploymentController) AddDeploymentContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	container := this.Ctx.Input.Param(":container")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	envVar := make([]corev1.EnvVar, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	d, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := d.Spec.Template.Spec

	newPodSpec, err := addPodSpecContainerEnv(podSpec, container, envVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	d.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, deployment, false)
	this.normalReturn("ok")
}

// DeleteDeploymentContainerEnv
// @Title Deployment
// @Description   Deployment Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param container path string true "容器"
// @Param env path string true "环境变量"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/container/:container/env/:env [Delete]
func (this *DeploymentController) DeleteDeploymentContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	container := this.Ctx.Input.Param(":container")
	env := this.Ctx.Input.Param(":env")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	d, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := d.Spec.Template.Spec

	newPodSpec, err := deletePodSpecContainerEnv(podSpec, container, env)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	d.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, deployment, false)
	this.normalReturn("ok")

}

// DeploymentContainerEnv
// @Title Deployment
// @Description   更新Deployment Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/container/:container/env [Put]
func (this *DeploymentController) UpdateDeploymentContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	container := this.Ctx.Input.Param(":container")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	envVar := make([]corev1.EnvVar, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	d, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := d.Spec.Template.Spec
	newPodSpec, err := updatePodSpecContainerEnv(podSpec, container, envVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	d.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, deployment, false)
	this.normalReturn("ok")

}

// GetDeploymentContainerVolume
// @Title Deployment
// @Description   Deployment Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/volume [Get]
func (this *DeploymentController) GetDeploymentVolumes() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	r, err := pi.GetRuntime()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	//vols := getSpecVolume(r.Deployment.Spec.Template.Spec)
	vols := getSpecVolumeAndVolumeMounts(r.Deployment.Spec.Template.Spec)

	this.normalReturn(vols)
}

// DeploymentContainerVolume
// @Title Deployment
// @Description   新增Deployment Container volume
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/volume [Post]
func (this *DeploymentController) AddDeploymentContainerSpecVolume() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	//	volumeVar := make([]corev1.VolumeMount, 0)
	volumeVar := VolumeAndVolumeMounts{}
	volumeVar.CMounts = make([]ContainerVolumeMount, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &volumeVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	d, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	//	podSpec := old.Spec.Template.Spec
	podSpec := d.Spec.Template.Spec

	newPodSpec, err := addVolumeAndContaienrVolumeMounts(podSpec, volumeVar)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	//old.Spec.Template.Spec = newPodSpec
	d.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, deployment, false)
	this.normalReturn("ok")

}

// DeleteDeploymentContainerVolume
// @Title Deployment
// @Description   Deployment Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Param volume path string true "卷"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/volume/:volume [Delete]
func (this *DeploymentController) DeleteDeploymentContainerSpecVolume() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	volume := this.Ctx.Input.Param(":volume")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	d, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := d.Spec.Template.Spec

	newPodSpec, err := deleteVolumeAndContaienrVolumeMounts(podSpec, volume)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	d.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(d)
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, deployment, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, deployment, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, deployment, false)
	this.normalReturn("ok")

}

// GetDeploymentServices
// @Title Deployment
// @Description   Deployment 服务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "部署"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/services [Get]
func (this *DeploymentController) GetDeploymentServices() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	v, err := pk.Controller.GetObject(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDeploymentInterface(v)

	services, err := pi.GetServices()
	if err != nil {
		this.errReturn(err, 500)
		return

	}

	this.normalReturn(services)

}
