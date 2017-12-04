package cluster

const (
	ActionDelete ActionType = "delete"
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
)

//用于避免和各资源的回环引用
var (
	PodEventChan                   = make(chan Event, 32)
	ServiceEventChan               = make(chan Event, 32)
	EndpointEventChan              = make(chan Event, 32)
	ConfigMapEventChan             = make(chan Event, 32)
	ReplicationControllerEventChan = make(chan Event, 32)
	ServiceAccountEventChan        = make(chan Event, 32)
	SecretEventChan                = make(chan Event, 32)

	DeploymentEventChan = make(chan Event, 32)
	ReplicaSetEventChan = make(chan Event, 32)
	RelicaSetEventChan  = make(chan Event, 32)
	DaemonSetEventChan  = make(chan Event, 32)
	IngressEventChan    = make(chan Event, 32)

	StatefulSetEventChan = make(chan Event, 32)
	CronJobEventChan     = make(chan Event, 32)
	JobEventChan         = make(chan Event, 32)
	HPAEventChan         = make(chan Event, 32)
)

type ActionType string

type Event struct {
	Group      string
	Workspace  string
	Name       string
	Action     ActionType
	FromUfleet bool //表明该资源由用户直接通过ufleet去创建的
}
