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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:CronJobController"],
		beego.ControllerComments{
			Method: "ListCronJobs",
			Router: `/group/:group/workspace/:workspace`,
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
			Method: "ListJobs",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:PodController"],
		beego.ControllerComments{
			Method: "ListPods",
			Router: `/group/:group/workspace/:workspace`,
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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:SecretController"],
		beego.ControllerComments{
			Method: "ListSecrets",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceAccountController"],
		beego.ControllerComments{
			Method: "ListServiceAccounts",
			Router: `/group/:group/workspace/:workspace`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:ServiceController"],
		beego.ControllerComments{
			Method: "ListServices",
			Router: `/group/:group/workspace/:workspace`,
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
