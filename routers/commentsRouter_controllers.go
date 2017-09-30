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

	beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"] = append(beego.GlobalControllerRouter["ufleet-deploy/controllers:TestController"],
		beego.ControllerComments{
			Method: "CheckDevMode",
			Router: `/devmode`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

}
