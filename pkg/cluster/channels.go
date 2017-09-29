package cluster

const (
	ActionDelete ActionType = "delete"
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
)

//用于避免和各资源的回环引用
var (
	PodEventChan            = make(chan Event)
	ServiceEventChan        = make(chan Event)
	EndpointEventChan       = make(chan Event)
	ConfigMapEventChan      = make(chan Event)
	ServiceAccountEventChan = make(chan Event)
	SecretEventChan         = make(chan Event)

	DeploymentEventChan = make(chan Event)
	DaemonSetEventChan  = make(chan Event)
	IngressEventChan    = make(chan Event)

	StatefulSetEventChan = make(chan Event)
	CronJobEventChan     = make(chan Event)
	JobEventChan         = make(chan Event)
)

type ActionType string

type Event struct {
	Group      string
	Workspace  string
	Name       string
	Action     ActionType
	FromUfleet bool //表明该资源由用户直接通过ufleet去创建的
}
