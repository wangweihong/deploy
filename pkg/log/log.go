package log

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/astaxie/beego"
)

func debugPrintFunc(err interface{}) string {
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

/*
  a := "bbb"
	uerr.PrintAndReturnError("aaa", a)
	uerr.PrintAndReturnError("%v---%v", a, "aaa")
	e := fmt.Errorf("aaa")
	uerr.PrintAndReturnError(e, "bbb")
	输出如下:

2017/08/09 10:47:06 [D] File(/home/wwh/kiongf/go/src/debug/test.go) Line(34) Func(main.main): aaabbb
2017/08/09 10:47:06 [D] File(/home/wwh/kiongf/go/src/debug/test.go) Line(35) Func(main.main): bbb---aaa
2017/08/09 10:47:06 [D] File(/home/wwh/kiongf/go/src/debug/test.go) Line(37) Func(main.main): aaabbb

*/
func ErrorPrint(data interface{}, a ...interface{}) error {
	switch err := data.(type) {
	case error:
		if err == nil {
			return nil
		}

		e := err
		for _, v := range a {
			e = fmt.Errorf("%v%v", e, v)
		}
		beego.Error(debugPrintFunc(e.Error()))
		return e

	case string:

		var e error
		str := fmt.Sprintf(err, a...)

		e = fmt.Errorf(str)

		//检测是否格式化出错
		//https://golang.org/pkg/fmt/
		if strings.Contains(e.Error(), "%!") {
			e = nil

			for _, v := range a {
				e = fmt.Errorf("%v%v", err, v)
			}
		}
		beego.Error(debugPrintFunc(e.Error()))
		return e

	default:

		var e error
		for _, v := range a {
			e = fmt.Errorf("%v%v", e, v)
		}
		beego.Error(debugPrintFunc(e.Error()))
		return e
	}
	return nil
}

func DebugPrint(data interface{}, a ...interface{}) error {

	switch err := data.(type) {
	case error:
		if err == nil {
			return nil
		}

		e := err
		for _, v := range a {
			e = fmt.Errorf("%v%v", e, v)
		}
		beego.Debug(debugPrintFunc(e.Error()))

		return e
	case string:

		var e error
		str := fmt.Sprintf(err, a...)

		e = fmt.Errorf(str)

		//检测是否格式化出错
		//https://golang.org/pkg/fmt/
		if strings.Contains(e.Error(), "%!") {
			e = nil

			for _, v := range a {
				e = fmt.Errorf("%v%v", err, v)
			}
		}
		beego.Debug(debugPrintFunc(e.Error()))
		return e

	default:

		var e error
		e = fmt.Errorf("%v", data)
		for _, v := range a {
			e = fmt.Errorf("%v%v", e, v)
		}
		beego.Debug(debugPrintFunc(e))

		return e
	}

	return nil
}
