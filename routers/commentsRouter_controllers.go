package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:AppController"],
		beego.ControllerComments{
			Method: "NewApp",
			Router: `/`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "ListConfigMaps",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "ListGroupConfigMaps",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "DeleteConfigMap",
			Router: `/:configmap/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ConfigMapController"],
		beego.ControllerComments{
			Method: "GetConfigMapTemplate",
			Router: `/:configmap/group/:group/workspace/:workspace/template`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListCronJobs",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListGroupCronJobs",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DaemonSetController"],
		beego.ControllerComments{
			Method: "ListDaemonSets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:DeploymentController"],
		beego.ControllerComments{
			Method: "ListDeployments",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:EndpointController"],
		beego.ControllerComments{
			Method: "ListEndpoints",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:IngressController"],
		beego.ControllerComments{
			Method: "ListIngresss",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:JobController"],
		beego.ControllerComments{
			Method: "ListGroupJobs",
			Router: `/group/:group/workspace/:workspace`,
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
			Method: "DeleteJob",
			Router: `/:job/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListPods",
			Router: `/group/:group/workspace/:workspace`,
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
			Method: "CreatePod",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListGroupPods",
			Router: `/groups`,
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ReplicationControllerController"],
		beego.ControllerComments{
			Method: "ListGroupReplicationControllers",
			Router: `/group/:group/workspace/:workspace`,
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
			Method: "DeleteReplicationController",
			Router: `/:replicationcontroller/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListSecrets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListGroupSecrets",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListServiceAccounts",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListGroupServiceAccounts",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListServices",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListGroupServices",
			Router: `/groups`,
			AllowHTTPMethods: []string{"Post"},
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:StatefulSetController"],
		beego.ControllerComments{
			Method: "ListStatefulSets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"],
		beego.ControllerComments{
			Method: "CheckDevMode",
			Router: `/devmode`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

}
