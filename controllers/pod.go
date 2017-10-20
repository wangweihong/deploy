package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/pod"
	//	"ufleet-deploy/util/user"
)

type PodController struct {
	baseController
}

type PodState struct {
	Name       string    `json:"name"`
	User       string    `json:"user"`
	App        string    `json:"app"`
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
	ps.App = pod.AppStack

	status, err := pi.GetStatus()
	if err != nil {
		ps.Reason = err.Error()
		return ps
	}
	ps.Status = *status
	if ps.CreateTime == 0 {
		ps.CreateTime = ps.Status.StartTime
	}

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

	fmt.Println(this.Ctx.Request.RequestURI)
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
func (this *PodController) ListGroupsPods() {

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

// ListGroupPods
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *PodController) ListGroupPods() {

	group := this.Ctx.Input.Param(":group")

	pis, err := pk.Controller.ListGroup(group)
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

// CreatePod
// @Title Pod
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *PodController) CreatePod() {

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

	err := pk.Controller.Delete(group, workspace, pod, resource.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// UpdatePod
// @Title Pod
// @Description  更新容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace [Put]
func (this *PodController) UpdatePod() {

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	/*
		ui := user.NewUserClient(token)
		ui.GetUserName()
	*/

	err := pk.Controller.Update(group, workspace, pod, this.Ctx.Input.RequestBody)
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

// GetPodContainers
// @Title Pod
// @Description   Pod container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/event [Get]
func (this *PodController) GetPodEvent() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	pi, err := pk.Controller.Get(group, workspace, pod)
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

// GetPodContainers
// @Title Pod
// @Description   Pod container terminal
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/container/:container/terminal [Get]
func (this *PodController) GetPodTerminal() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	logs, err := pi.Terminal(c)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(logs)
}

// GetPodContainerSpec
// @Title Pod
// @Description   Pod Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/container/:container/spec [Get]
func (this *PodController) GetPodContainerSpec() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	container := this.Ctx.Input.Param(":container")

	pi, err := pk.Controller.Get(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	stat, err := pi.GetStatus()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	for _, v := range stat.ContainerSpecs {
		if v.Name == container {
			this.normalReturn(v)
			return
		}
	}

	err = fmt.Errorf("container not found")

	this.errReturn(err, 500)
}
