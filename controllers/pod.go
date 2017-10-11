package controllers

import (
	"encoding/json"
	"fmt"
	pk "ufleet-deploy/pkg/resource/pod"
)

type PodController struct {
	baseController
}

type PodState struct {
	Name       string    `json:"name"`
	User       string    `json:"user"`
	Workspace  string    `json:"workspace"`
	CreateTime int64     `json:"createtime"`
	Reason     string    `json:"reason"`
	Message    string    `json:"message"`
	Status     pk.Status `json:"status"`
}

func GetPodState(pi pk.PodInterface) PodState {
	var ps PodState
	pod := pi.Info()
	ps.Name = pod.Name
	ps.User = pod.User
	ps.CreateTime = pod.CreateTime
	ps.Workspace = pod.Workspace

	status, err := pi.GetStatus()
	if err != nil {
		ps.Reason = err.Error()
		return ps
	}
	ps.Status = *status
	if ps.CreateTime == 0 {
		ps.CreateTime = ps.Status.StartTime
	}
	/*
		runtime, err := pi.GetRuntime()
		if err != nil {
			ps.Reason = err.Error()
			return ps
		}

			if runtime.Pod.Status.StartTime != nil {
				ps.Status.StartTime = runtime.Pod.Status.StartTime.Unix()
			}
			ps.Status.State = string(runtime.Pod.Status.Phase)
			ps.Status.HostIP = runtime.Pod.Status.HostIP
			ps.Status.IP = runtime.Pod.Generation
	*/

	return ps

}

// ListPods
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *PodController) ListPods() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pss := make([]PodState, 0)
	for _, v := range pis {
		ps := GetPodState(v)
		pss = append(pss, ps)
	}

	this.normalReturn(pss)
}

// GetPod
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace [Get]
func (this *PodController) GetPod() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	s, err := pi.GetStatus()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(s)
}

// ListGroupsPods
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *PodController) ListGroupPods() {

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

	pis := make([]pk.PodInterface, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	pss := make([]PodState, 0)
	for _, v := range pis {
		ps := GetPodState(v)
		pss = append(pss, ps)
	}

	this.normalReturn(pss)
}

// DeletePod
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace [Delete]
func (this *PodController) DeletePod() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	err := pk.Controller.Delete(group, workspace, pod, pk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// GetPodTemplate
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/template [Get]
func (this *PodController) GetPodTemplate() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	pi, err := pk.Controller.Get(group, workspace, pod)
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

// GetPodContainers
// @Title Pod
// @Description   Pod Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/containers [Get]
func (this *PodController) GetPodContainers() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	infor, err := pi.GetRuntime()
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	cs := make([]string, 0)

	for _, v := range infor.Pod.Spec.Containers {
		cs = append(cs, v.Name)
	}

	this.normalReturn(cs)
}

// GetPodContainers
// @Title Pod
// @Description   Pod container log
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/container/:container/log [Get]
func (this *PodController) GetPodLog() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	logs, err := pi.Log(c)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(logs)
}

// GetPodContainers
// @Title Pod
// @Description   Pod container stat
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/container/:container/stat [Get]
func (this *PodController) GetPodStat() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	logs, err := pi.Stat(c)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(logs)
}
