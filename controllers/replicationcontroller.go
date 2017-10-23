package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/resource"
	jk "ufleet-deploy/pkg/resource/replicationcontroller"
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := jk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]jk.ReplicationController, 0)
	jss := make([]jk.Status, 0)

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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	rc := this.Ctx.Input.Param(":rc")

	pi, err := jk.Controller.Get(group, workspace, rc)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]jk.ReplicationController, 0)
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

	pis := make([]jk.ReplicationControllerInterface, 0)
	for _, v := range groups {
		tmp, err := jk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//replicationcontrollers := make([]jk.ReplicationController, 0)
	jss := make([]jk.Status, 0)
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
	group := this.Ctx.Input.Param(":group")
	pis, err := jk.Controller.ListGroup(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]jk.ReplicationController, 0)
	jss := make([]jk.Status, 0)
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
	err := jk.Controller.Create(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

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

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := jk.Controller.Update(group, workspace, replicationcontroller, this.Ctx.Input.RequestBody)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

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

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	replicasStr := this.Ctx.Input.Param(":replicas")

	ri, err := jk.Controller.Get(group, workspace, replicationcontroller)
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	err := jk.Controller.Delete(group, workspace, replicationcontroller, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	pi, err := jk.Controller.Get(group, workspace, replicationcontroller)
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

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	pi, err := jk.Controller.Get(group, workspace, replicationcontroller)
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
