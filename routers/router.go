// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"ufleet-deploy/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1/deploy",
		beego.NSNamespace("/test",
			beego.NSInclude(
				&controllers.TestController{},
			),
		),
		beego.NSNamespace("/stack",
			beego.NSInclude(
				&controllers.AppController{},
			),
		),
		beego.NSNamespace("/program",
			beego.NSInclude(&controllers.ProgramController{}),
		),
		beego.NSNamespace("/pod",
			beego.NSInclude(
				&controllers.PodController{},
			),
		),
		beego.NSNamespace("/service",
			beego.NSInclude(
				&controllers.ServiceController{},
			),
		),
		beego.NSNamespace("/secret",
			beego.NSInclude(
				&controllers.SecretController{},
			),
		),
		beego.NSNamespace("/configmap",
			beego.NSInclude(
				&controllers.ConfigMapController{},
			),
		),
		beego.NSNamespace("/replicationcontroller",
			beego.NSInclude(
				&controllers.ReplicationControllerController{},
			),
		),
		beego.NSNamespace("/serviceaccount",
			beego.NSInclude(
				&controllers.ServiceAccountController{},
			),
		),
		beego.NSNamespace("/endpoint",
			beego.NSInclude(
				&controllers.EndpointController{},
			),
		),
		beego.NSNamespace("/deployment",
			beego.NSInclude(
				&controllers.DeploymentController{},
			),
		),
		beego.NSNamespace("/daemonset",
			beego.NSInclude(
				&controllers.DaemonSetController{},
			),
		),
		beego.NSNamespace("/ingress",
			beego.NSInclude(
				&controllers.IngressController{},
			),
		),
		beego.NSNamespace("/statefulset",
			beego.NSInclude(
				&controllers.StatefulSetController{},
			),
		),
		beego.NSNamespace("/job",
			beego.NSInclude(
				&controllers.JobController{},
			),
		),
		beego.NSNamespace("/cronjob",
			beego.NSInclude(
				&controllers.CronJobController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
