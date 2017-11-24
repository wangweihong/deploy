package system

import (
	"strconv"
	"time"
)

// NewAuditClient 新建Audit
func NewAuditClient() AuditLog {
	var audit AuditLog
	etcd := NewEtcd3Client()
	audit.etcd = etcd
	return audit
}

// AuditLog 审计日志
type AuditLog struct {
	ID       string
	etcd     Etcd3Client
	Operator string // 操作用户名称
	Operate  string // 操作内容
	Object   string // 操作对象
	Level    string // INFO ERROR
	Module   string // 模块名称， user,uflow,deploy 等，
	Time     string
	UnixTime int64
}

// Create 创建审计日志
func (a *AuditLog) Create() error {
	a.Time = time.Now().Format("2006-01-02 15:04:05")
	a.UnixTime = time.Now().UnixNano()
	a.ID = a.Module + "_" + strconv.FormatInt(a.UnixTime, 10)
	_, err := a.etcd.Set("/ufleet/auditlog/"+a.ID, a)
	return err
}
