package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/user"
	//	"ufleet-deploy/util/user"
	corev1 "k8s.io/api/core/v1"
)

type PodController struct {
	baseController
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
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
		v, _ := pk.GetPodInterface(j)
		ps := v.GetStatus()
		pss = append(pss, *ps)
	}

	this.normalReturn(pss)
}

// GetGroupPodCounts
// @Title Pod
// @Description  统计组Pod
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/count [Get]
func (this *PodController) GetGroupPodCount() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")

	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	var failedCount int
	pods := make([]*corev1.Pod, 0)
	for _, v := range pis {
		pi, _ := pk.GetPodInterface(v)
		r, err := pi.GetRuntime()
		if err != nil {
			failedCount += 1
			//			this.errReturn(err, 500)
			//			return
			continue
		}

		pods = append(pods, r.Pod)
	}
	c := resource.GetPodsCount(pods, failedCount)

	this.normalReturn(c)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	fmt.Println(this.Ctx.Request.RequestURI)
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

	s := pi.GetStatus()

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	groups := make([]string, 0)
	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.errReturn(err, 500)
		return
	}

	err = json.Unmarshal(this.Ctx.Input.RequestBody, &groups)
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
		v, _ := pk.GetPodInterface(j)
		ps := v.GetStatus()
		pss = append(pss, *ps)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
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
		v, _ := pk.GetPodInterface(j)
		ps := v.GetStatus()
		pss = append(pss, *ps)
	}

	this.normalReturn(pss)
}

// ListGroupPods
// @Title Pod
// @Description   Pod
// @Param Token header string true 'Token'
// @Success 201 {string} create success!
// @Failure 500
// @router /allgroup/count [Get]
func (this *PodController) GetAllGroupPodsCount() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	groups := pk.Controller.ListGroups()

	pods := make([]*corev1.Pod, 0)
	var failedCount int
	for _, v := range groups {
		pis, err := pk.Controller.ListGroupObject(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}

		for _, j := range pis {
			v, _ := pk.GetPodInterface(j)
			r, err := v.GetRuntime()
			if err != nil {
				failedCount += 1
				//				this.errReturn(err, 500)
				//				return
				continue //不返回错误
			}
			pods = append(pods, r.Pod)
		}

	}

	c := resource.GetPodsCount(pods, failedCount)
	this.normalReturn(c)
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
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
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
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	err = pk.Controller.DeleteObject(group, workspace, pod, resource.DeleteOption{})
	if err != nil {
		this.audit(token, pod, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, pod, false)
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
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, pod, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, pod, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, pod, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, pod, false)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetPodInterface(v)
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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	c := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetPodInterface(v)

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
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

	stat := pi.GetStatus()
	if stat.Reason != "" {
		this.errReturn(fmt.Errorf(stat.Reason), 500)
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

// GetPodContainerEnv
// @Title Pod
// @Description   Pod Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/container/:container/env [Get]
func (this *PodController) GetPodContainerSpecEnv() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

	stat := pi.GetStatus()
	if stat.Reason != "" {
		this.errReturn(fmt.Errorf(stat.Reason), 500)
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

// GetPodServices
// @Title Pod
// @Description   Pod 服务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:pod/group/:group/workspace/:workspace/services [Get]
func (this *PodController) GetPodServices() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	pod := this.Ctx.Input.Param(":pod")

	v, err := pk.Controller.GetObject(group, workspace, pod)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetPodInterface(v)

	services, err := pi.GetServices()
	if err != nil {
		this.errReturn(err, 500)
		return

	}

	this.normalReturn(services)

}
