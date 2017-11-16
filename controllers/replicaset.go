package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/replicaset"
	"ufleet-deploy/pkg/user"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type ReplicaSetController struct {
	baseController
}

// ListReplicaSets
// @Title ReplicaSet
// @Description   ReplicaSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *ReplicaSetController) ListGroupWorkspaceReplicaSets() {
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
	//replicasets := make([]pk.ReplicaSet, 0)
	jss := make([]pk.Status, 0)

	for _, j := range pis {
		v, _ := pk.GetReplicaSetInterface(j)
		js := v.GetStatus()

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// GetReplicaSet
// @Title ReplicaSet
// @Description   ReplicaSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param rc path string true "副本控制器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:rc/group/:group/workspace/:workspace [Get]
func (this *ReplicaSetController) GetReplicaSet() {
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
	//replicasets := make([]pk.ReplicaSet, 0)
	pi, _ := pk.GetReplicaSetInterface(v)

	js := pi.GetStatus()

	this.normalReturn(js)
}

// ListGroupsReplicaSets
// @Title ReplicaSet
// @Description   ReplicaSet
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *ReplicaSetController) ListGroupsReplicaSets() {
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
	//replicasets := make([]pk.ReplicaSet, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetReplicaSetInterface(j)
		js := v.GetStatus()

		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// ListGroupsReplicaSets
// @Title ReplicaSet
// @Description   ReplicaSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *ReplicaSetController) ListGroupReplicaSets() {
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
	//replicasets := make([]pk.ReplicaSet, 0)
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetReplicaSetInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}

	this.normalReturn(jss)
}

// CreateReplicaSet
// @Title ReplicaSet
// @Description  创建副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *ReplicaSetController) CreateReplicaSet() {
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
		this.audit(token, "", true)
		err := fmt.Errorf("must commit resource json/yaml data")
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

// UpdateReplicaSet
// @Title ReplicaSet
// @Description  更新副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本控制器"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace [Put]
func (this *ReplicaSetController) UpdateReplicaSet() {
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
	replicaset := this.Ctx.Input.Param(":replicaset")

	if this.Ctx.Input.RequestBody == nil {
		this.audit(token, "", true)
		err := fmt.Errorf("must commit resource json/yaml data")
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, replicaset, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
}

// ScaleReplicaSet
// @Title ReplicaSet
// @Description  扩容副本控制器
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本控制器"
// @Param replicas path string true "副本数"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/replicas/:replicas [Put]
func (this *ReplicaSetController) ScaleReplicaSet() {
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
	replicaset := this.Ctx.Input.Param(":replicaset")
	replicasStr := this.Ctx.Input.Param(":replicas")

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	ri, _ := pk.GetReplicaSetInterface(v)

	err = ri.Scale(int(replicas))
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")

}

// DeleteReplicaSet
// @Title ReplicaSet
// @Description  DeleteeReplicaSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "任务名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace [Delete]
func (this *ReplicaSetController) DeleteReplicaSet() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	err := pk.Controller.DeleteObject(group, workspace, replicaset, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
}

// GetReplicaSetContainerEvents
// @Title ReplicaSet
// @Description   ReplicaSet container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/event [Get]
func (this *ReplicaSetController) GetReplicaSetEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}

// GetReplicaSetTemplate
// @Title ReplicaSet
// @Description   ReplicaSet
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "容器组"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/template [Get]
func (this *ReplicaSetController) GetReplicaSetTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetDeploymentContainerEnv
// @Title Deployment
// @Description   Deployment Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param container path string true "容器"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/container/:container/env [Get]
func (this *ReplicaSetController) GetReplicaSetContainerSpecEnv() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

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

// ReplicaSetContainerEnv
// @Title ReplicaSet
// @Description   新增ReplicaSet Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/container/:container/env [Post]
func (this *ReplicaSetController) AddReplicaSetContainerSpecEnv() {
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
	replicaset := this.Ctx.Input.Param(":replicaset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.ReplicaSet
	var containerFound bool
	var containerIndex int
	podSpec := old.Spec.Template.Spec

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			containerIndex = k

			break
		}
	}

	if !containerFound {
		err = fmt.Errorf("container not found")
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}

	podSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar...)

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	err = pk.Controller.UpdateObject(group, workspace, replicaset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// DeleteReplicaSetContainerEnv
// @Title ReplicaSet
// @Description   ReplicaSet Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param container path string true "容器"
// @Param env path string true "环境变量"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/container/:container/env/:env [Delete]
func (this *ReplicaSetController) DeleteReplicaSetContainerSpecEnv() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		this.audit(token, "", true)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")
	container := this.Ctx.Input.Param(":container")
	env := this.Ctx.Input.Param(":env")
	log.DebugPrint(env)

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.ReplicaSet
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

	err = pk.Controller.UpdateObject(group, workspace, replicaset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// ReplicaSetContainerEnv
// @Title ReplicaSet
// @Description   更新ReplicaSet Container env
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param container path string true "容器"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/container/:container/env [Put]
func (this *ReplicaSetController) UpdateReplicaSetContainerSpecEnv() {
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
	replicaset := this.Ctx.Input.Param(":replicaset")
	container := this.Ctx.Input.Param(":container")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.ReplicaSet
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

	err = pk.Controller.UpdateObject(group, workspace, replicaset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// GetReplicaSetContainerVolume
// @Title ReplicaSet
// @Description   ReplicaSet Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/volume [Get]
func (this *ReplicaSetController) GetReplicaSetVolume() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	r, err := pi.GetRuntime()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	vols := getSpecVolume(r.ReplicaSet.Spec.Template.Spec)

	this.normalReturn(vols)
}

// ReplicaSetContainerVolume
// @Title ReplicaSet
// @Description   新增ReplicaSet Container volume
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param body body string true "更新内容"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/volume [Post]
func (this *ReplicaSetController) AddReplicaSetContainerSpecVolume() {
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

	volumeVar := VolumeAndVolumeMounts{}
	volumeVar.CMounts = make([]ContainerVolumeMount, 0)
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &volumeVar)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.ReplicaSet
	podSpec := old.Spec.Template.Spec

	newPodSpec, err := addVolumeAndContaienrVolumeMounts(podSpec, volumeVar)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	old.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
	}

	err = pk.Controller.UpdateObject(group, workspace, replicaset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// DeleteReplicaSetContainerVolume
// @Title ReplicaSet
// @Description   ReplicaSet Containers
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Param volume path string true "卷挂载"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/volume/:volume [Delete]
func (this *ReplicaSetController) DeleteReplicaSetContainerSpecVolume() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		this.audit(token, "", true)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")
	volume := this.Ctx.Input.Param(":volume")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	runtime, err := pi.GetRuntime()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old := runtime.ReplicaSet
	podSpec := old.Spec.Template.Spec

	newPodSpec, err := deleteVolumeAndContaienrVolumeMounts(podSpec, volume)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	old.Spec.Template.Spec = newPodSpec

	byteContent, err := json.Marshal(old)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)

	}

	err = pk.Controller.UpdateObject(group, workspace, replicaset, byteContent, resource.UpdateOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)
	this.normalReturn("ok")

}

// GetReplicaSetServices
// @Title ReplicaSet
// @Description   ReplicaSet 服务
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param replicaset path string true "副本集"
// @Success 201 {string} create success!
// @Failure 500
// @router /:replicaset/group/:group/workspace/:workspace/services [Get]
func (this *ReplicaSetController) GetReplicaSetServices() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	replicaset := this.Ctx.Input.Param(":replicaset")

	v, err := pk.Controller.GetObject(group, workspace, replicaset)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetReplicaSetInterface(v)

	services, err := pi.GetServices()
	if err != nil {
		this.errReturn(err, 500)
		return

	}

	this.normalReturn(services)

}
