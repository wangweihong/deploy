package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "NewApp",
			Router: `/:app/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "DeleteApp",
			Router: `/:app/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "RecreateApp",
			Router: `/:app/group/:group/workspace/:workspace/recreate`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "UpdateApp",
			Router: `/:app/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "GetApp",
			Router: `/:app/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "GetAppTemplate",
			Router: `/:app/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "GetAppGroupCounts",
			Router: `/group/:group/counts`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "GetAppGroupsCounts",
			Router: `/groups/counts`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "ListGroupApp",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "GetClusterAppsCount",
			Router: `/apps/cluster`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "AppAddResource",
			Router: `/:app/group/:group/workspace/:workspace/resources`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "AppRemoveResource",
			Router: `/:app/group/:group/workspace/:workspace/kind/:kind/resource/:resource`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceConfigMaps",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "ListGroupsConfigMaps",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "ListGroupConfigMaps",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "CreateConfigMap",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "CreateConfigMapCustom",
			Router: `/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "UpdateConfigMapCustom",
			Router: `/:configmap/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "DeleteConfigMap",
			Router: `/:configmap/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "UpdateConfigMap",
			Router: `/:configmap/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "GetConfigMapTemplate",
			Router: `/:configmap/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "GetConfigMapEvent",
			Router: `/:configmap/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "GetConfigMapReferenceObject",
			Router: `/:configmap/group/:group/workspace/:workspace/reference`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceCronJobs",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "GetCronJob",
			Router: `/:cronjob/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListGroupsCronJobs",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListGroupCronJobs",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "CreateCronJob",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "UpdateCronJob",
			Router: `/:cronjob/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "DeleteCronJob",
			Router: `/:cronjob/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "GetCronJobTemplate",
			Router: `/:cronjob/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "GetCronJobEvent",
			Router: `/:cronjob/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "SuspendOrResumeCronJob",
			Router: `/:cronjob/group/:group/workspace/:workspace/suspendandresume`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "ListDaemonSets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "ListGroupDaemonSets",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSet",
			Router: `/:daemonset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "CreateDaemonSet",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "UpdateDaemonSet",
			Router: `/:daemonset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "UpdateDaemonSetCustom",
			Router: `/:daemonset/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "DeleteDaemonSet",
			Router: `/:daemonset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetEvent",
			Router: `/:daemonset/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetTemplate",
			Router: `/:daemonset/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetRevisions",
			Router: `/:daemonset/group/:group/workspace/:workspace/revisions`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "RollBackDaemonSet",
			Router: `/:daemonset/group/:group/workspace/:workspace/revision/:revision`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetContainerSpecEnv",
			Router: `/:daemonset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "AddDaemonSetContainerSpecEnv",
			Router: `/:daemonset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "DeleteDaemonSetContainerSpecEnv",
			Router: `/:daemonset/group/:group/workspace/:workspace/container/:container/env/:env`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "UpdateDaemonSetContainerSpecEnv",
			Router: `/:daemonset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetVolumes",
			Router: `/:daemonset/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "AddDaemonSetContainerSpecVolume",
			Router: `/:daemonset/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "DeleteDaemonSetContainerSpecVolume",
			Router: `/:daemonset/group/:group/workspace/:workspace/volume/:volume`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "UpdateDaemonSetContainerSpecVolume",
			Router: `/:daemonset/group/:group/workspace/:workspace/container/:container/volume`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "GetDaemonSetServices",
			Router: `/:daemonset/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceDeployments",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "ListGroupDeployments",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "CreateDeployment",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "UpdateDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "UpdateDeploymentCustom",
			Router: `/:deployment/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "DeleteDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentEvent",
			Router: `/:deployment/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "ScaleDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace/replicas/:replicas`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "ScaleDeploymentIncrement",
			Router: `/:deployment/group/:group/workspace/:workspace/increment/:increment`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentTemplate",
			Router: `/:deployment/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentReplicaSet",
			Router: `/:deployment/group/:group/workspace/:workspace/replicasets`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentRevisions",
			Router: `/:deployment/group/:group/workspace/:workspace/revisions`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "RollBackDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace/revision/:revision`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetHPA",
			Router: `/:deployment/group/:group/workspace/:workspace/hpa`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "StartHPA",
			Router: `/:deployment/group/:group/workspace/:workspace/hpa`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "RollBackResumeOrPauseDeployment",
			Router: `/:deployment/group/:group/workspace/:workspace/resumeorpause`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentContainerSpecEnv",
			Router: `/:deployment/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "AddDeploymentContainerSpecEnv",
			Router: `/:deployment/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "DeleteDeploymentContainerSpecEnv",
			Router: `/:deployment/group/:group/workspace/:workspace/container/:container/env/:env`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "UpdateDeploymentContainerSpecEnv",
			Router: `/:deployment/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentVolumes",
			Router: `/:deployment/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "AddDeploymentContainerSpecVolume",
			Router: `/:deployment/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "DeleteDeploymentContainerSpecVolume",
			Router: `/:deployment/group/:group/workspace/:workspace/volume/:volume`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "GetDeploymentServices",
			Router: `/:deployment/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "ListEndpoints",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "GetGroupWorkspaceEndpoint",
			Router: `/:endpoint/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "ListGroupsEndpoints",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "ListGroupEndpoints",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "CreateEndpoint",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "UpdateEndpoint",
			Router: `/:endpoint/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "DeleteEndpoint",
			Router: `/:endpoint/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "GetEndpointEvent",
			Router: `/:endpoint/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "ListIngresss",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "GetGroupWorkspaceIngress",
			Router: `/:ingress/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "ListGroupsIngresss",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "ListGroupIngresss",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "CreateIngress",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "UpdateIngress",
			Router: `/:ingress/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "DeleteIngress",
			Router: `/:ingress/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "GetIngressEvent",
			Router: `/:ingress/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "GetIngressTemplate",
			Router: `/:ingress/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "GetGroupWorkspaceIngressServices",
			Router: `/:ingress/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceJobs",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "GetJob",
			Router: `/:job/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "ListGroupsJobs",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "ListGroupJobs",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "CreateJob",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "UpdateJob",
			Router: `/:job/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "DeleteJob",
			Router: `/:job/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "GetJobTemplate",
			Router: `/:job/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "GetJobEvent",
			Router: `/:job/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListPods",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetGroupPodCount",
			Router: `/group/:group/count`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPod",
			Router: `/:pod/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListGroupsPods",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListGroupPods",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetAllGroupPodsCount",
			Router: `/allgroup/count`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "CreatePod",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "DeletePod",
			Router: `/:pod/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "UpdatePod",
			Router: `/:pod/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodTemplate",
			Router: `/:pod/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodContainers",
			Router: `/:pod/group/:group/workspace/:workspace/containers`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodLog",
			Router: `/:pod/group/:group/workspace/:workspace/container/:container/log`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodStat",
			Router: `/:pod/group/:group/workspace/:workspace/container/:container/stat`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodEvent",
			Router: `/:pod/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodTerminal",
			Router: `/:pod/group/:group/workspace/:workspace/container/:container/terminal`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodContainerSpec",
			Router: `/:pod/group/:group/workspace/:workspace/container/:container/spec`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodContainerSpecEnv",
			Router: `/:pod/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "GetPodServices",
			Router: `/:pod/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ProgramController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ProgramController"],
		beego.ControllerComments{
			Method: "GetVersion",
			Router: `/version`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ProgramController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ProgramController"],
		beego.ControllerComments{
			Method: "GetLogs",
			Router: `/logs`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceReplicaSets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSet",
			Router: `/:rc/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "ListGroupsReplicaSets",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "ListGroupReplicaSets",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "CreateReplicaSet",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "UpdateReplicaSet",
			Router: `/:replicaset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "ScaleReplicaSet",
			Router: `/:replicaset/group/:group/workspace/:workspace/replicas/:replicas`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "DeleteReplicaSet",
			Router: `/:replicaset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSetEvent",
			Router: `/:replicaset/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSetTemplate",
			Router: `/:replicaset/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSetContainerSpecEnv",
			Router: `/:replicaset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "AddReplicaSetContainerSpecEnv",
			Router: `/:replicaset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "DeleteReplicaSetContainerSpecEnv",
			Router: `/:replicaset/group/:group/workspace/:workspace/container/:container/env/:env`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "UpdateReplicaSetContainerSpecEnv",
			Router: `/:replicaset/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSetVolume",
			Router: `/:replicaset/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "AddReplicaSetContainerSpecVolume",
			Router: `/:replicaset/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "DeleteReplicaSetContainerSpecVolume",
			Router: `/:replicaset/group/:group/workspace/:workspace/volume/:volume`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicaSetController"],
		beego.ControllerComments{
			Method: "GetReplicaSetServices",
			Router: `/:replicaset/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceReplicationControllers",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationController",
			Router: `/:rc/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "ListGroupsReplicationControllers",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "ListGroupReplicationControllers",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "CreateReplicationController",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "UpdateReplicationController",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "ScaleReplicationController",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/replicas/:replicas`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "DeleteReplicationController",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationControllerEvent",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationControllerTemplate",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationControllerContainerSpecEnv",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "AddReplicationControllerContainerSpecEnv",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "DeleteReplicationControllerContainerSpecEnv",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/container/:container/env/:env`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "UpdateReplicationControllerContainerSpecEnv",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/container/:container/env`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationControllerVolume",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "AddReplicationControllerVolume",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/volume`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "DeleteReplicationControllerVolume",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/volume/:volume`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "UpdateReplicationControllerContainerSpecVolume",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/container/:container/volume`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "GetReplicationControllerServices",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListSecrets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListGroupsSecrets",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListGroupSecrets",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "CreateSecret",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "CreateSecretCustom",
			Router: `/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "UpdateSecret",
			Router: `/:secret/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "DeleteSecret",
			Router: `/:secret/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "GetSecretTemplate",
			Router: `/:secret/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "GetSecretEvent",
			Router: `/:secret/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "GetSecretReferenceObject",
			Router: `/:secret/group/:group/workspace/:workspace/reference`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListServiceAccounts",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListGroupsServiceAccounts",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListGroupServiceAccounts",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "CreateServiceAccount",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "CreateServiceAccountCustom",
			Router: `/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "UpdateServiceAccount",
			Router: `/:serviceaccount/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "UpdateServiceAccountCustom",
			Router: `/:serviceaccount/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "DeleteServiceAccount",
			Router: `/:serviceaccount/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "GetServiceAccountTemplate",
			Router: `/:serviceaccount/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "GetServiceAccountEvent",
			Router: `/:serviceaccount/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "GetServiceAccountReferenceObject",
			Router: `/:serviceaccount/group/:group/workspace/:workspace/reference`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListGroupWorkspaceServices",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "GetGroupWorkspaceService",
			Router: `/:service/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListGroupsServices",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListGroupServices",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "CreateService",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "UpdateService",
			Router: `/:service/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "DeleteService",
			Router: `/:service/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "GetServiceTemplate",
			Router: `/:service/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "GetServiceEvent",
			Router: `/:service/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "GetServiceReferenceObject",
			Router: `/:service/group/:group/workspace/:workspace/reference`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "GetServiceReferenceIngresses",
			Router: `/:service/group/:group/workspace/:workspace/ingresses`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "ListStatefulSets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "GetGroupWorkspaceStatefulSet",
			Router: `/:statefulset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "ListGroupsStatefulSets",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "ListGroupStatefulSets",
			Router: `/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "CreateStatefulSet",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "UpdateStatefulSet",
			Router: `/:statefulset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "UpdateStatefulSetCustom",
			Router: `/:statefulset/group/:group/workspace/:workspace/custom`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "DeleteStatefulSet",
			Router: `/:statefulset/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "GetStatefulSetEvent",
			Router: `/:statefulset/group/:group/workspace/:workspace/event`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "GetStatefulSetServices",
			Router: `/:statefulset/group/:group/workspace/:workspace/services`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "GetStatefulSetTemplate",
			Router: `/:statefulset/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"],
		beego.ControllerComments{
			Method: "CheckDevMode",
			Router: `/devmode`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

}
