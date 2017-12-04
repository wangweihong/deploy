package sign

const (
	//用于标记该资源是ufleet创建的
	SignFromUfleetKey   = "resource-come-from"
	SignFromUfleetValue = "com.appsoar.ufleet"

	SignUfleetAppKey             = "com.appsoar.ufleet.app"
	SignUfleetAutoScaleSupported = "com.appsoar.ufleet.autoscale" //指定哪些deploymnet支持他行伸缩
	SignUfleetDeployment         = "com.appsoar.ufleet.deploy"    //在pod指定哪些pod属于它
)
