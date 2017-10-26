package audit

import (
	"log"
	"os"
	"path/filepath"
	"time"
	"ufleet-deploy/pkg/kv"
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
	/*
		go func(aobj AuditObj) {
			auditChan <- aobj
		}(aobj)
	*/
	auditHandler.Audit(aobj)
}

func NewAuditHandler() AuditHandler {
	httpHandler := initHttpAuditHandler()
	return httpHandler

}

/*
func (l *AuditLogger) Audit() {
	for {

		aobj := <-auditChan
		switch aobj.Level {
		case AuditLevelInfo, AuditLevelError:
		default:
			log.ErrorPrint("invalid audit level ", aobj.Level)
			continue
		}

		l.Printf("%v-%02v-%02v %02v:%02v:%02v %v %v %v %v %v %v\n", aobj.Time.Year(), int(aobj.Time.Month()), aobj.Time.Day(), aobj.Time.Hour(), aobj.Time.Minute(), aobj.Time.Second(), aobj.Time.Unix(), aobj.Level, aobj.Operator, aobj.Operate, aobj.Object, aobj.ObjectName)
		//不刷新,数据不会写入到文件中
		err := l.Sync()
		if err != nil {
			log.ErrorPrint(err.Error())
		}
	}
}
*/

type fileAuditHandler struct {
	*log.Logger
	*os.File
}

func (l *fileAuditHandler) Audit(aobj AuditObj) {
	l.Printf("%v-%02v-%02v %02v:%02v:%02v %v %v %v %v %v %v\n", aobj.Time.Year(), int(aobj.Time.Month()), aobj.Time.Day(), aobj.Time.Hour(), aobj.Time.Minute(), aobj.Time.Second(), aobj.Time.Unix(), aobj.Level, aobj.Operator, aobj.Operate, aobj.Object, aobj.ObjectName)
	err := l.Sync()
	if err != nil {
		dlog.ErrorPrint(err.Error())
	}
}
func initFileAuditHandler() (*fileAuditHandler, error) {

	if err := os.MkdirAll(filepath.Dir(auditLogFilePath), 0755); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}

	auditLogFile, err := os.OpenFile(auditLogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := log.New(auditLogFile, "", 0)
	auditLogger := &fileAuditHandler{logger, auditLogFile}
	return auditLogger, nil

}

type httpAuditHandler struct {
	//	auditClient
}

func (l *httpAuditHandler) Audit(aobj AuditObj) {
	auditClient, err := ufleetSystem.NewAuditClient([]string{kv.GetKVStoreAddr()})
	if err != nil {
		dlog.ErrorPrint(err)
	}

	auditClient.Level = string(aobj.Level)
	auditClient.Object = aobj.ObjectName
	auditClient.Operator = aobj.Operator
	auditClient.Operate = aobj.Operate
	auditClient.Module = aobj.Object

	err = auditClient.Create()
	if err != nil {
		dlog.ErrorPrint(err.Error())
	}
}

func initHttpAuditHandler() *httpAuditHandler {
	return &httpAuditHandler{}
}

func init() {
	auditHandler = NewAuditHandler()
	/*
		if err := os.MkdirAll(filepath.Dir(auditLogFilePath), 0755); err != nil {
			if !os.IsExist(err) {
				panic(err.Error())
			}
		}

		auditLogFile, err := os.OpenFile(auditLogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err.Error())
		}

		go func() {
			for {
				select {
				case <-time.Tick(5 * time.Second):
					auditLogFile.Sync()
				}
			}
		}()

		logger := log.New(auditLogFile, "", 0)
		auditLogger := AuditLogger{logger, auditLogFile}
		go auditLogger.Audit()
	*/
}
