package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	_ "ufleet-deploy/init"
	_ "ufleet-deploy/routers"

	"ufleet-deploy/mode"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	} else {
		if err := os.Mkdir(filepath.Dir("./logs/"), 0755); err != nil {
			if !os.IsExist(err) {
				panic(err.Error())
			}
		}
		logs.EnableFuncCallDepth(true)
		logs.SetLogger(logs.AdapterFile, `{"filename":"./logs/deploy.log", "perm":"0664"}`)
	}

	Trap()
	beego.Run()
}

func Trap() {
	c := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGUSR1, syscall.SIGUSR2}
	signal.Notify(c, signals...)
	go func() {
		for sig := range c {
			go func(sig os.Signal) {
				switch sig {
				case syscall.SIGUSR1:
					mode.TurnProductModeOff()
					beego.BConfig.WebConfig.DirectoryIndex = true
					beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
					beego.SetLevel(logs.LevelDebug)
				case syscall.SIGUSR2:
					mode.TurnProductModeOn()
					beego.BConfig.WebConfig.DirectoryIndex = false
					beego.BConfig.WebConfig.StaticDir["/swagger"] = ""
					beego.SetLevel(logs.LevelInfo)
				}
			}(sig)
		}
	}()
}
