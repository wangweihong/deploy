package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/replicationcontroller"
	"ufleet-deploy/pkg/user"
)

type ReplicationControllerController struct {
	baseController
}

// ListReplicationControllers
// @Title ReplicationController
// @Description   ReplicationController
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ReplicationControllerController) ListGroupWorkspaceReplicationControllers() {
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
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)

	for _, v := range pis {
		js := v.GetStatus()

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetReplicationController
// @Title ReplicationController
// @Description   ReplicationController
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param rc path string true "副本控制器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:rc/group/:group/workspace/:workspace [Get]
func (this *ReplicationControllerController) GetReplicationController() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	rc := this.Ctx.Input.Param(":rc")

	pi, err := pk.Controller.Get(group, workspace, rc)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	v := pi

	js := v.GetStatus()

	this.normalReturn(js)
}

// ListGroupsReplicationControllers
// @Title ReplicationController
// @Description   ReplicationController
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ReplicationControllerController) ListGroupsReplicationControllers() {
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

	pis := make([]pk.ReplicationControllerInterface, 0)
	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupsReplicationControllers
// @Title ReplicationController
// @Description   ReplicationController
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *ReplicationControllerController) ListGroupReplicationControllers() {
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
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)
	for _, v := range pis {
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateReplicationController
// @Title ReplicationController
// @Description  创建副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ReplicationControllerController) CreateReplicationController() {
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

// UpdateReplicationController
// @Title ReplicationController
// @Description  更新副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace [Put]
func (this *ReplicationControllerController) UpdateReplicationController() {
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
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

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

	err := pk.Controller.Update(group, workspace, replicationcontroller, this.Ctx.Input.RequestBody)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// ScaleReplicationController
// @Title ReplicationController
// @Description  扩容副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param replicas path string true "副本数"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/replicas/:replicas [Put]
func (this *ReplicationControllerController) ScaleReplicationController() {
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
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	replicasStr := this.Ctx.Input.Param(":replicas")

	ri, err := pk.Controller.Get(group, workspace, replicationcontroller)
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

// DeleteReplicationController
// @Title ReplicationController
// @Description  DeleteeReplicationController
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "任务名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace [Delete]
func (this *ReplicationControllerController) DeleteReplicationController() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	err := pk.Controller.Delete(group, workspace, replicationcontroller, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// GetReplicationControllerContainerEvents
// @Title ReplicationController
// @Description   ReplicationController container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/event [Get]
func (this *ReplicationControllerController) GetReplicationControllerEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	pi, err := pk.Controller.Get(group, workspace, replicationcontroller)
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

// GetReplicationControllerTemplate
// @Title ReplicationController
// @Description   ReplicationController
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/template [Get]
func (this *ReplicationControllerController) GetReplicationControllerTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	pi, err := pk.Controller.Get(group, workspace, replicationcontroller)
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
