package audit

import (
	"log"
	"os"
	"time"
	dlog "ufleet-deploy/pkg/log"

	ufleetSystem "go-ufleetutil/system"
)

type AuditLevel string

const (
	auditLogFilePath            = "/var/log/ufleet/deploy.log"
	AuditLevelInfo   AuditLevel = "INFO"
	AuditLevelError  AuditLevel = "ERROR"
)

var (
	//	auditChan = make(chan AuditObj)
	auditHandler AuditHandler
)

type AuditObj struct {
	Time       time.Time  `json:"time"`
	Operator   string     `json:"operator"`
	Operate    string     `json:"operate"`
	Object     string     `json:"object"`
	ObjectName string     `json:"objectName"`
	Level      AuditLevel `json:"level"`
}

type AuditHandler interface {
	Audit(aobj AuditObj)
}

type AuditLogger struct {
	*log.Logger
	*os.File
}

func Audit(aobj AuditObj) {
	auditHandler.Audit(aobj)
}

func NewAuditHandler() AuditHandler {
	httpHandler := initHttpAuditHandler()
	return httpHandler

}

type httpAuditHandler struct {
	//	auditClient
}

func (l *httpAuditHandler) Audit(aobj AuditObj) {
	auditClient := ufleetSystem.NewAuditClient()

	auditClient.Level = string(aobj.Level)
	auditClient.Object = aobj.Object + " " + aobj.ObjectName
	auditClient.Operator = aobj.Operator
	auditClient.Operate = aobj.Operate
	auditClient.Module = "deploy"

	err := auditClient.Create()
	if err != nil {
		dlog.ErrorPrint(err.Error())
	}
}

func initHttpAuditHandler() *httpAuditHandler {
	return &httpAuditHandler{}
}

func init() {
	auditHandler = NewAuditHandler()
}
