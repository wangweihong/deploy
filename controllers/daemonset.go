package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/daemonset"
	"ufleet-deploy/pkg/user"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type DaemonSetController struct {
	baseController
}

// ListDaemonSets
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *DaemonSetController) ListDaemonSets() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetDaemonSetInterface(j)
		js := v.GetStatus()

		jss = append(jss, *js)
	}
	this.normalReturn(jss)
}

// ListGroupDaemonSets
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *DaemonSetController) ListGroupDaemonSets() {
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
	for _, j := range pis {
		v, _ := pk.GetDaemonSetInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)

}

// GetDaemonSets
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区名"
// @Param daemonset path string true "守护进程"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace [Get]
func (this *DaemonSetController) GetDaemonSet() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	pi, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	v, _ := pk.GetDaemonSetInterface(pi)
	js := v.GetStatus()

	this.normalReturn(js)

}

// CreateDaemonSet
// @Title DaemonSet
// @Description  创建容器组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *DaemonSetController) CreateDaemonSet() {
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

	err = pk.Controller.CreateObject(group, workspace, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateDaemonSet
// @Title DaemonSet
// @Description  更新daemonset组
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "daemonset组"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace [Put]
func (this *DaemonSetController) UpdateDaemonSet() {
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
	daemonset := this.Ctx.Input.Param(":daemonset")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, daemonset, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteDaemonSet
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace [Delete]
func (this *DaemonSetController) DeleteDaemonSet() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	err := pk.Controller.DeleteObject(group, workspace, daemonset, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
}

// GetDaemonSetContainerEvents
// @Title DaemonSet
// @Description   DaemonSet container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/event [Get]
func (this *DaemonSetController) GetDaemonSetEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}

// GetDaemonSetTemplate
// @Title DaemonSet
// @Description   DaemonSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/template [Get]
func (this *DaemonSetController) GetDaemonSetTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)
	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetDaemonSetRevisionsAndDecribe
// @Title DaemonSet
// @Description   DaemonSet 版本描述
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "守护进程"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/revisions [Get]
func (this *DaemonSetController) GetDaemonSetRevisions() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetDaemonSetInterface(v)
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
// @Title DaemonSet
// @Description   DaemonSet回滚
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "守护进程"
// @Param revision path string true "版本"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/revision/:revision [Put]
func (this *DaemonSetController) RollBackDaemonSet() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")
	revision := this.Ctx.Input.Param(":revision")

	toRevision, err := strconv.ParseInt(revision, 10, 64)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)

	result, err := pi.Rollback(toRevision)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return

	}

	this.audit(token, "", false)
	this.normalReturn(*result)
}

// GetDaemonSetContainerEnv
// @Title DaemonSet
// @Description   DaemonSet Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param pod path string true "容器组"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/container/:container/env [Get]
func (this *DaemonSetController) GetDaemonSetContainerSpecEnv() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)

	stat := pi.GetStatus()
	if err != nil {
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

// DaemonSetContainerEnv
// @Title DaemonSet
// @Description   新增DaemonSet Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "守护进程"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/container/:container/env [Post]
func (this *DaemonSetController) AddDaemonSetContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		this.audit(token, "", true)
		return
	}

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var envVar corev1.EnvVar
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.DaemonSet
	var containerFound bool
	var envFound bool
	var containerIndex int
	podSpec := old.Spec.Template.Spec

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			containerIndex = k

			for _, j := range v.Env {
				if j.Name == envVar.Name {
					envFound = true
					break
				}
			}
			break
		}
	}

	if !containerFound {
		err = fmt.Errorf("container not found")
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}

	if envFound {
		err = fmt.Errorf("env has exist")
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}

	podSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar)

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	err = pk.Controller.UpdateObject(group, workspace, daemonset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// DeleteDaemonSetContainerEnv
// @Title DaemonSet
// @Description   DaemonSet Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "守护进程"
// @Param container path string true "容器"
// @Param env path string true "环境变量"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/container/:container/env/:env [Delete]
func (this *DaemonSetController) DeleteDaemonSetContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		this.audit(token, "", true)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")
	container := this.Ctx.Input.Param(":container")
	env := this.Ctx.Input.Param(":env")
	log.DebugPrint(env)

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.DaemonSet
	var containerFound bool
	var envFound bool
	podSpec := old.Spec.Template.Spec

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			for i, j := range v.Env {
				log.DebugPrint(j.Name)
				if j.Name == env {
					podSpec.Containers[k].Env = append(podSpec.Containers[k].Env[:i], podSpec.Containers[k].Env[i+1:]...)
					envFound = true
					log.DebugPrint("found env", v.Name)
					break
				}
			}
			break
		}
	}

	if !containerFound {
		err = fmt.Errorf("container not found")
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}
	if !envFound {
		err = fmt.Errorf("env not found")
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	err = pk.Controller.UpdateObject(group, workspace, daemonset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// DaemonSetContainerEnv
// @Title DaemonSet
// @Description   更新DaemonSet Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param daemonset path string true "守护进程"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:daemonset/group/:group/workspace/:workspace/container/:container/env [Put]
func (this *DaemonSetController) UpdateDaemonSetContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		this.audit(token, "", true)
		return
	}

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	envVar := make([]corev1.EnvVar, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	daemonset := this.Ctx.Input.Param(":daemonset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, daemonset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetDaemonSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.DaemonSet
	var containerFound bool
	podSpec := old.Spec.Template.Spec

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			podSpec.Containers[k].Env = envVar

			break
		}
	}

	if !containerFound {
		err = fmt.Errorf("container not found")
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	err = pk.Controller.UpdateObject(group, workspace, daemonset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}
