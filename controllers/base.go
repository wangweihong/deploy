package controllers

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	uaudit "ufleet-deploy/pkg/audit"
	user "ufleet-deploy/pkg/user"

	"github.com/astaxie/beego"
)

type baseController struct {
	beego.Controller
}

type ErrStruct struct {
	Err  string `json:"error_msg"`
	Code int    `json:"error_code"`
}

func debugPrintFunc(err string) string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)

	if n == 0 {
		return "n/a"
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	file, line := fun.FileLine(fpcs[0])
	return fmt.Sprintf("File(%v) Line(%v) Func(%v): %v", file, line, fun.Name(), err)
}

func (this *baseController) errReturn(data interface{}, statusCode int) {
	var errStruct ErrStruct
	errStruct.Code = statusCode
	switch v := data.(type) {
	case string:
		errStruct.Err = v
	case error:
		errStruct.Err = v.Error()
	case ErrStruct:
		errStruct = v

	}

	debugErr := fmt.Errorf("RequestIP:%v,Error:%v", this.Ctx.Request.RemoteAddr, errStruct.Err)
	beego.Error(debugPrintFunc(debugErr.Error()))
	//	uerr.PrintAndReturnError(err)

	this.Ctx.Output.SetStatus(statusCode)
	this.Data["json"] = errStruct

	this.ServeJSON()
}

func (this *baseController) normalReturn(data interface{}, statusCode ...int) {
	this.Data["json"] = data

	if len(statusCode) != 0 {
		this.Ctx.Output.SetStatus(statusCode[0])
	}

	this.ServeJSON()
}

func getRouteControllerName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)

	if n == 0 {
		return "n/a"
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	sl := strings.Split(fun.Name(), ".")
	return sl[len(sl)-1]

}

//没有ability的为GET方法,不进行权限判定
func getRouteControllerAbility(funName string) *ability {

	abi, ok := abilityMap[funName]
	if !ok {
		return nil
	}
	return &abi
}

//TODO:向用户模块验证token是否有效
func checkTokenAvailable(token string) (bool, error) {
	ui := user.NewUserClient(token)
	_, err := ui.GetUserName()
	if err != nil {
		return false, err
	}
	return true, nil
}

//TODO: 添加获取token用户角色
func checkUserAbilityAllowed(token string, abi string) (bool, error) {
	return true, nil
}

func (this *baseController) checkRouteControllerAbility() error {
	//调试!
	//	return nil

	token := this.Ctx.Request.Header.Get("token")

	ok, err := checkTokenAvailable(token)
	if err != nil {
		beego.Error(err.Error())
		return errTokenInvalid
	}
	if !ok {
		return errTokenInvalid
	}

	fun := getRouteControllerName()
	abi := getRouteControllerAbility(fun)
	if abi == nil {
		return nil
	}

	allowed, err := checkUserAbilityAllowed(token, abi.object)
	if err != nil {
		return err
	}

	if !allowed {
		return errPermessionDenied
	}

	return nil
}

func (this *baseController) abilityErrorReturn(err error) {
	if err != nil {
		if err == errTokenInvalid {
			this.errReturn(err, 401)
			return
		}
		if err == errPermessionDenied {
			this.errReturn(err, 403)
			return
		}
		this.errReturn(err, 500)
		return
	}
}

func (this *baseController) audit(token string, objectName string, meetError bool) {
	var ad uaudit.AuditObj

	ui := user.NewUserClient(token)
	username, err := ui.GetUserName()
	if err != nil {
		beego.Error(fmt.Sprintf("audit fail for can not get user %v", err))
		return
	}

	fpcs := make([]uintptr, 4)
	n := runtime.Callers(2, fpcs)

	if n == 0 {
		beego.Error("audit fail for can not get router's name")
		return
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		beego.Error("audit fail for can not get router's name")
		//			return "n/a"
		return
	}

	sl := strings.Split(fun.Name(), ".")
	funName := sl[len(sl)-1]
	audit, ok := auditMap[funName]
	if !ok {
		beego.Warn("ignore invalid audit router name ")
		return
	}

	//	ad.Time = time.Now().Format("2006-01-02 15:04:05")
	ad.Time = time.Now()
	ad.Object = audit.object
	ad.Operate = audit.operate
	ad.Operator = username
	ad.ObjectName = objectName
	if meetError {
		ad.Level = uaudit.AuditLevelError
	} else {
		ad.Level = uaudit.AuditLevelInfo
	}

	uaudit.Audit(ad)
	return
}
