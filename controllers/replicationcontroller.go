package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/replicationcontroller"
	"ufleet-deploy/pkg/user"

	corev1 "k8s.io/api/core/v1"
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

	pis, err := pk.Controller.ListGroupWorkspaceObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)

	for _, j := range pis {
		v, _ := pk.GetReplicationControllerInterface(j)
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

	v, err := pk.Controller.GetObject(group, workspace, rc)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	pi, _ := pk.GetReplicationControllerInterface(v)

	js := pi.GetStatus()

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

	pis := make([]resource.Object, 0)
	for _, v := range groups {
		tmp, err := pk.Controller.ListGroupObject(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetReplicationControllerInterface(j)
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
	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//replicationcontrollers := make([]pk.ReplicationController, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetReplicationControllerInterface(j)
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
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, replicationcontroller, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, replicationcontroller, false)
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
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	replicasStr := this.Ctx.Input.Param(":replicas")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	ri, _ := pk.GetReplicationControllerInterface(v)
	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = ri.Scale(int(replicas))
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, replicationcontroller, false)
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
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	err := pk.Controller.DeleteObject(group, workspace, replicationcontroller, resource.DeleteOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, replicationcontroller, false)
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

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

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

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	pi, _ := pk.GetReplicationControllerInterface(v)
	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetReplicationControllerContainerEnv
// @Title ReplicationController
// @Description   ReplicationController Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/container/:container/env [Get]
func (this *ReplicationControllerController) GetReplicationControllerContainerSpecEnv() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

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

// ReplicationControllerContainerEnv
// @Title ReplicationController
// @Description   新增ReplicationController Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/container/:container/env [Post]
func (this *ReplicationControllerController) AddReplicationControllerContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	container := this.Ctx.Input.Param(":container")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	envVar := make([]corev1.EnvVar, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	r, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	var containerFound bool
	var containerIndex int
	podSpec := r.Spec.Template.Spec

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			containerIndex = k

			break
		}
	}

	if !containerFound {
		err = fmt.Errorf("container not found")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	podSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar...)

	byteContent, err := json.Marshal(r)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, replicationcontroller, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, replicationcontroller, false)
	this.normalReturn("ok")
}

// DeleteReplicationControllerContainerEnv
// @Title ReplicationController
// @Description   ReplicationController Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param container path string true "容器"
// @Param env path string true "环境变量"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/container/:container/env/:env [Delete]
func (this *ReplicationControllerController) DeleteReplicationControllerContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	container := this.Ctx.Input.Param(":container")
	env := this.Ctx.Input.Param(":env")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	old, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

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
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
	}
	if !envFound {
		err = fmt.Errorf("env not found")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, replicationcontroller, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, replicationcontroller, false)
	this.normalReturn("ok")

}

// ReplicationControllerContainerEnv
// @Title ReplicationController
// @Description   更新ReplicationController Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/container/:container/env [Put]
func (this *ReplicationControllerController) UpdateReplicationControllerContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	container := this.Ctx.Input.Param(":container")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	envVar := make([]corev1.EnvVar, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &envVar)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	old, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

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
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, replicationcontroller, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, replicationcontroller, false)
	this.normalReturn("ok")

}

// GetReplicationControllerContainerVolume
// @Title ReplicationController
// @Description   ReplicationController Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/volume [Get]
func (this *ReplicationControllerController) GetReplicationControllerVolume() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	r, err := pi.GetRuntime()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	vols := getSpecVolumeAndVolumeMounts(r.ReplicationController.Spec.Template.Spec)
	this.normalReturn(vols)
}

// ReplicationControllerContainerVolume
// @Title ReplicationController
// @Description   新增ReplicationController Container volume
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/volume [Post]
func (this *ReplicationControllerController) AddReplicationControllerVolume() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	var volumeVar VolumeAndVolumeMounts
	volumeVar.CMounts = make([]ContainerVolumeMount, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &volumeVar)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	old, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := old.Spec.Template.Spec

	newPodSpec, err := addVolumeAndContaienrVolumeMounts(podSpec, volumeVar)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	old.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, replicationcontroller, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, replicationcontroller, false)
	this.normalReturn("ok")

}

// DeleteReplicationControllerContainerVolume
// @Title ReplicationController
// @Description   ReplicationController Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Param volume path string true "卷挂载"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/volume/:volume [Delete]
func (this *ReplicationControllerController) DeleteReplicationControllerVolume() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")
	volume := this.Ctx.Input.Param(":volume")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	old, err := pi.GetRuntimeObjectCopy()
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	podSpec := old.Spec.Template.Spec

	newPodSpec, err := deleteVolumeAndContaienrVolumeMounts(podSpec, volume)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	old.Spec.Template.Spec = newPodSpec
	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}

	err = pk.Controller.UpdateObject(group, workspace, replicationcontroller, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, replicationcontroller, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, replicationcontroller, false)
	this.normalReturn("ok")

}

// GetReplicationControllerServices
// @Title ReplicationController
// @Description   ReplicationController 服务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicationcontroller path string true "副本控制器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicationcontroller/group/:group/workspace/:workspace/services [Get]
func (this *ReplicationControllerController) GetReplicationControllerServices() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicationcontroller := this.Ctx.Input.Param(":replicationcontroller")

	v, err := pk.Controller.GetObject(group, workspace, replicationcontroller)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicationControllerInterface(v)

	services, err := pi.GetServices()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(services)
}
