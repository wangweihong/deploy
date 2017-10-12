package controllers

import (
	"encoding/json"
	"fmt"
	pk "ufleet-deploy/pkg/resource/pod"
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
		js := &jk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			replicationcontroller := v.Info()
			js.Name = replicationcontroller.Name
			js.User = replicationcontroller.User
			js.Workspace = replicationcontroller.Workspace
			js.Group = replicationcontroller.Group
			js.Reason = err.Error()
			js.PodStatus = make([]pk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
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
		js := &jk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			replicationcontroller := v.Info()
			js.Name = replicationcontroller.Name
			js.User = replicationcontroller.User
			js.Workspace = replicationcontroller.Workspace
			js.Group = replicationcontroller.Group
			js.Reason = err.Error()
			js.PodStatus = make([]pk.Status, 0)
			jss = append(jss, *js)
			continue
		}

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
// @router /groups [Post]
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
		js := &jk.Status{}
		var err error
		js, err = v.GetStatus()
		if err != nil {
			replicationcontroller := v.Info()
			js.Name = replicationcontroller.Name
			js.User = replicationcontroller.User
			js.Workspace = replicationcontroller.Workspace
			js.Group = replicationcontroller.Group
			js.Reason = err.Error()
			js.PodStatus = make([]pk.Status, 0)
			jss = append(jss, *js)
			continue
		}

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
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

	err := jk.Controller.Delete(group, workspace, replicationcontroller, jk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
