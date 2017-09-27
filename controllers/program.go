package controllers

import (
	"io"
	"os"
	"ufleet-deploy/pkg/log"
)

type ProgramController struct {
	baseController
}

// GetVersion
// @Title
// @Description   当前版本
// @Success 200 {string}
// @Failure 500
// @router /version [Get]
func (this *ProgramController) GetVersion() {
	this.normalReturn("v1.4.0.153", 200)
}

// GetLogs
// @Title
// @Description 日志
// @Success 200 {string}
// @Failure 500
// @router /logs [Get]
func (this *ProgramController) GetLogs() {

	file, err := os.Open("./logs/deploy.log")
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	size := 1024 * 1024
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		log.DebugPrint(err)
		this.errReturn(err, 500)
		return
	}

	var start int64
	var buf []byte
	if stat.Size() <= int64(size) {
		start = 0
		buf = make([]byte, stat.Size())
	} else {
		start = stat.Size() - int64(size)
		buf = make([]byte, size)
	}
	_, err = file.ReadAt(buf, start)
	if err != nil {
		if err != io.EOF {
			log.DebugPrint(err)
			this.errReturn(err, 500)
			return
		}
	}

	this.Ctx.WriteString(string(buf))
}
