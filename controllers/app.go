package controllers

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/app"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource/cronjob"
	"ufleet-deploy/pkg/resource/daemonset"
	"ufleet-deploy/pkg/resource/deployment"
	"ufleet-deploy/pkg/resource/job"
	"ufleet-deploy/pkg/resource/pod"
	"ufleet-deploy/pkg/resource/replicaset"
	"ufleet-deploy/pkg/resource/replicationcontroller"
	"ufleet-deploy/pkg/resource/service"
	"ufleet-deploy/pkg/resource/statefulset"
	"ufleet-deploy/pkg/user"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type AppController struct {
	baseController
}

// newApps
// @Title 应用
// @Description   添加新的应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Post]
func (this *AppController) NewApp() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	ui := user.NewUserClient(token)
	who, err := ui.GetUserName()
	if err != nil {
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}
	var opt app.CreateOption
	opt.User = who

	err = app.Controller.NewApp(group, workspace, appName, this.Ctx.Input.RequestBody, opt)
	if err != nil {
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, appName, false)
	this.normalReturn("ok")
}

// deleteApps
// @Title 应用
// @Description   删除应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Delete]
func (this *AppController) DeleteApp() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	err := app.Controller.DeleteApp(group, workspace, appName, app.DeleteOption{})
	if err != nil {
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, appName, false)
	this.normalReturn("ok")
}

// App
// @Title 应用
// @Description  重建应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace/recreate [Put]
func (this *AppController) RecreateApp() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}
	log.DebugPrint(string(this.Ctx.Input.RequestBody))

	err = app.Controller.RecreateApp(group, workspace, appName, this.Ctx.Input.RequestBody, app.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.audit(token, appName, false)
	this.normalReturn("ok")
}

// App
// @Title 应用
// @Description  重建应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Put]
func (this *AppController) UpdateApp() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}

	err = app.Controller.UpdateApp(group, workspace, appName, this.Ctx.Input.RequestBody, app.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.audit(token, appName, false)
	this.normalReturn("ok")
}

// GetApp
// @Title 应用
// @Description   获取指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace [Get]
func (this *AppController) GetApp() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	ai, err := app.Controller.Get(group, workspace, appName)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(ai.Info())
}

// GetApp
// @Title 应用
// @Description   获取指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace/template [Get]
func (this *AppController) GetAppTemplate() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	ai, err := app.Controller.Get(group, workspace, appName)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	t, err := ai.GetTemplates()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	s := fmt.Sprintf("")
	for _, v := range t {
		s = fmt.Sprintf("%v\n---\n%v", s, v)
	}

	this.normalReturn(s)
}

// GetApp
// @Title 应用
// @Description   获取指定组应用统计
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/counts [Get]
func (this *AppController) GetAppGroupCounts() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")

	ais, err := app.Controller.List(group, app.ListOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(len(ais))
}

// GetApp
// @Title 应用
// @Description   获取所有组应用统计
// @Param Token header string true 'Token'
// @Success 201 {string} create success!
// @Failure 500
// @router /groups/counts [Get]
func (this *AppController) GetAppGroupsCounts() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	ais := app.Controller.ListGroupsApps()

	this.normalReturn(len(ais))
}

// ListGroupApp
// @Title 应用
// @Description   添加指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *AppController) ListGroupApp() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")

	ais, err := app.Controller.List(group, app.ListOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	statuses := make([]app.Status, 0)
	for _, v := range ais {
		statuses = append(statuses, v.GetStatus())
	}

	this.normalReturn(statuses)
}

// GetAppsFromCluster
// @Title 应用
// @Description   获取集群上的应用
// @Param Token header string true 'Token'
// @Param body body string true "集群信息"
// @Success 201 {string} create success!
// @Failure 500
// @router /apps/cluster [Post]
func (this *AppController) GetClusterAppsCount() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	wgs := struct {
		Clusterwgs []struct {
			Gws []struct {
				Group     string   `json:"group"`
				Workspace []string `json:"workspace"`
			} `json:"gws"`
			Cluster string `json:"cluster"`
		} `json:"clusterwgs"`
	}{}

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit group/workspace info")
		this.errReturn(err, 500)
		return
	}

	err = json.Unmarshal(this.Ctx.Input.RequestBody, &wgs)
	if err != nil {
		err = fmt.Errorf("try to unmarshal data \"%v\" fail for %v", string(this.Ctx.Input.RequestBody), err)
		this.errReturn(err, 500)
		return
	}

	can := make(map[string]int)

	for _, j := range wgs.Clusterwgs {
		for _, i := range j.Gws {
			for _, w := range i.Workspace {
				ais, err := app.Controller.ListGroupWorkspaceApps(i.Group, w)
				if err != nil {
					this.errReturn(err, 500)
					return
				}

				can[j.Cluster] += len(ais)
			}
		}

	}

	this.normalReturn(can)
}

// App
// @Title 应用
// @Description  应用添加资源
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace/resources [Put]
func (this *AppController) AppAddResource() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}

	err = app.Controller.AddAppResource(group, workspace, appName, this.Ctx.Input.RequestBody, app.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.audit(token, appName, false)
	this.normalReturn("ok")

}

// App
// @Title 应用
// @Description  应用删除资源
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Param kind path string true "资源类型"
// @Param resource path string true "资源名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace/kind/:kind/resource/:resource [Delete]
func (this *AppController) AppRemoveResource() {
	token := this.Ctx.Request.Header.Get("token")
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	appName := this.Ctx.Input.Param(":app")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	kind := this.Ctx.Input.Param(":kind")
	resource := this.Ctx.Input.Param(":resource")

	err = app.Controller.RemoveAppResource(group, workspace, appName, kind, resource)
	if err != nil {
		this.audit(token, appName, true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, appName, false)
	this.normalReturn("ok")

}

type ServiceAndContainerPort struct {
	Services      []ServiceDescribe       `json:"services"`
	ResourcePorts []ResourceContainerPort `json:"resourcePorts"`
}

type ServiceDescribe struct {
	Name  string               `json:"name"`
	Ports []corev1.ServicePort `json:"ports"`
}

type ResourceContainerPort struct {
	OwnerKind      string            `json:"ownerKind"`
	Owner          string            `json:"owner"`
	Selectors      map[string]string `json:"selectors"`
	ContainerPorts []ContainerPort   `json:"containerPorts"`
}

type ContainerPort struct {
	Container string                 `json:"container"`
	Ports     []corev1.ContainerPort `json:"ports"`
}

// GetAppServices
// @Title 应用
// @Description   添加指定应用
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param app path string true "栈名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:app/group/:group/workspace/:workspace/serviceAndContainerPort [Get]
func (this *AppController) GetAppServiceAndContainerPort() {
	err := this.checkRouteControllerAbility()
	if err != nil {
		this.abilityErrorReturn(err)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	stack := this.Ctx.Input.Param(":app")

	ai, err := app.Controller.Get(group, workspace, stack)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	as := ai.GetStatus()
	if as.Reason != "" {
		this.errReturn(err, 500)
		return
	}
	var sc ServiceAndContainerPort
	sc.Services = make([]ServiceDescribe, 0)
	sc.ResourcePorts = make([]ResourceContainerPort, 0)

	for _, os := range as.Statues {

		switch s := os.(type) {
		case *cronjob.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Seletors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *daemonset.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selectors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *deployment.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selectors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *job.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selector
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *pod.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Labels
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *replicaset.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selectors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *replicationcontroller.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selectors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *statefulset.Status:
			var rcp ResourceContainerPort
			rcp.ContainerPorts = make([]ContainerPort, 0)
			rcp.OwnerKind = s.Kind
			rcp.Owner = s.Name
			rcp.Selectors = s.Selectors
			for _, v := range s.ContainerSpecs {
				var cp ContainerPort
				cp.Container = v.Name
				cp.Ports = make([]corev1.ContainerPort, 0)
				cp.Ports = append(cp.Ports, v.Ports...)
				rcp.ContainerPorts = append(rcp.ContainerPorts, cp)
			}
			sc.ResourcePorts = append(sc.ResourcePorts, rcp)
		case *service.Status:
			var sd ServiceDescribe
			sd.Name = s.Name
			sd.Ports = make([]corev1.ServicePort, 0)
			sd.Ports = append(sd.Ports, s.Ports...)
			sc.Services = append(sc.Services, sd)

		default:
			continue
		}
	}
	this.normalReturn(sc)
}
