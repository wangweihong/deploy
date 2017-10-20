package controllers

import (
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/deployment"
	jk "ufleet-deploy/pkg/resource/pod"
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := &pk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			deployment := v.Info()
			js.Name = deployment.Name
			js.User = deployment.User
			js.Workspace = deployment.Workspace
			js.Group = deployment.Group
			js.Reason = err.Error()
			js.PodStatus = make([]jk.Status, 0)
			jss = append(jss, *js)
			continue
		}
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

	group := this.Ctx.Input.Param(":group")

	pis, err := pk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := &pk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			deployment := v.Info()
			js.Name = deployment.Name
			js.User = deployment.User
			js.Workspace = deployment.Workspace
			js.Group = deployment.Group
			js.Reason = err.Error()
			js.PodStatus = make([]jk.Status, 0)
			jss = append(jss, *js)
			continue
		}
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	pi, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v := pi
	js := &pk.Status{}
	js, err = v.GetStatus()
	if err != nil {
		deployment := v.Info()
		js.Name = deployment.Name
		js.User = deployment.User
		js.Workspace = deployment.Workspace
		js.Group = deployment.Group
		js.Reason = err.Error()
		js.PodStatus = make([]jk.Status, 0)
	}

	this.normalReturn(js)

}

// CreateDeployment
// @Title Deployment
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *DeploymentController) CreateDeployment() {

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
	err := pk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

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
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// DeleteDeployment
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace [Delete]
func (this *DeploymentController) DeleteDeployment() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")

	err := pk.Controller.Delete(group, workspace, deployment, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetDeploymentContainerEvents
// @Title Deployment
// @Description   Deployment container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/event [Get]
func (this *DeploymentController) GetDeploymentEvent() {

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

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	deployment := this.Ctx.Input.Param(":deployment")
	replicasStr := this.Ctx.Input.Param(":replicas")

	ri, err := pk.Controller.Get(group, workspace, deployment)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	err = ri.Scale(int(replicas))
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")

}

// GetDeploymentTemplate
// @Title Deployment
// @Description   Deployment
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param deployment path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:deployment/group/:group/workspace/:workspace/template [Get]
func (this *DeploymentController) GetDeploymentTemplate() {

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
