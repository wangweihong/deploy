package models

type PodRuntime struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	IP        string `json:"ip"`
	HostIP    string `json:"hostip"`
	StartTime string `json:"starttime"`
}

type CreateOption struct {
	Comment string `json:"comment"`
	Data    string `json:"data"`
}

/*
var (
	ServiceStateNormal  = "normal"
	ServiceStateWaiting = "waiting"
	ServiceStateError   = "error"

	AppStateNormal  = "normal"
	AppStateWaiting = "waiting"
	AppStateError   = "error"

	PodStateNormal    = "normal"
	PodStateWaiting   = "waiting"
	PodStateError     = "error"
	PodStateCompleted = "complete"

	TaskStateNormal  = "normal"
	TaskStateWaiting = "waiting"
	TaskStateError   = "error"

	JobStateNormal   = "normal"
	JobStateWaiting  = "waiting"
	JobStateError    = "error"
	JobStateComplete = "complete"
)

type GroupAppsState struct {
	States    []AppState `json:"appstates"`
	GroupName string     `json:"group"`
}

type AppState struct {
	AppName     string         `json:"app"`
	State       string         `json:"state"`
	ServiceNum  int            `json:"servicenum"`
	PodNum      int            `json:"podnum"`
	Services    []ServiceState `json:"services"`
	CreateTime  int64          `json:"createtime"`
	Group       string         `json:"group"`
	User        string         `json:"user"`
	Workspace   string         `json:"workspace"`
	Stop        bool           `json:"stop"`
	ClusterIP   string         `json:"clusterip"`
	ClusterName string         `json:"clusterName"`
}

//用于取代AppState,下一版本
type AppServicesState struct {
	AppName     string             `json:"app"`
	State       string             `json:"state"`
	ServiceNum  int                `json:"servicenum"`
	PodNum      int                `json:"podnum"`
	Services    []ServicePodsState `json:"services"`
	CreateTime  int64              `json:"createtime"`
	Group       string             `json:"group"`
	User        string             `json:"user"`
	Workspace   string             `json:"workspace"`
	Stop        bool               `json:"stop"`
	ClusterName string             `json:"clusterName"`
	ClusterIP   string             `json:"clusterip"`
}

type PortMap struct {
	//	ServicePort int32  `json:"servicePort"`
	NodePort int32 `json:"nodePort"`
	//	Protocol    string `json:"protocol"`
}

type ServiceCondition struct {
	LastUpdateTime string `json:"lastUpdateTime"`
	Reason         string `json:"reason"`
	Message        string `json:"message"`
}

type ServiceState struct {
	Kind           string             `json:"kind"`
	DesiredPodNum  int                `json:"desiredPodNum"`
	Reason         string             `json:"reason"`
	ServiceName    string             `json:"service"`
	State          string             `json:"state"`
	Images         []string           `json:"images"`
	PodNum         int                `json:"podnum"`
	App            string             `json:"app"`
	Ports          []int32            `json:"ports"`
	CreateTime     int64              `json:"createtime"`
	User           string             `json:"user"`
	Conditions     []ServiceCondition `json:"conditions"`
	Workspace      string             `json:"workspace"`
	ClusterIP      string             `json:"clusterip"`
	Stop           bool               `json:"stop"`
	CurrentVersion string             `json:"currentversion"`
	ClusterName    string             `json:"clusterName"`
}

type PodState struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	IP         string `json:"ip"`
	PodIP      string `json:"podip"`
	Running    int    `json:"running"`
	Total      int    `json:"total"`
	Message    string `json:"message"`
	Reason     string `json:"reason"`
	CreateTime int64  `json:"createtime"`
	Uid        string `json:"uid"`
}

type ServicePodsState struct {
	Service ServiceState `json:"service"`
	Pods    []PodState   `json:"pods"`
}

type ServiceTemplate struct {
	Name string          `json:"name"`
	Desc json.RawMessage `json:"desc"`
}

type AppServicePodCount struct {
	AppNormalNum  int `json:"appNormalNum"`
	AppWaitingNum int `json:"appWaitingNum"`
	AppErrorNum   int `json:"appErrorNum"`

	ServiceNormalNum  int `json:"serviceNormalNum"`
	ServiceWaitingNum int `json:"serviceWaitingNum"`
	ServiceErrorNum   int `json:"serviceErrorNum"`

	PodNormalNum  int `json:"podNormalNum"`
	PodWaitingNum int `json:"podWaitingNum"`
	PodErrorNum   int `json:"podErrorNum"`
}

type Container struct {
	Name string `json:"name"` //pod中的容器名
	ID   string `json:"id"`   //pod中的containerID
}

type ContainerImage struct {
	Container string `json:"container"`
	Image     string `json:"image"`
}
type ServiceUpgradeOption struct {
	//ObjectVersion string          `json:"objectVersion,omitempty"`
	Describe json.RawMessage `json:"desc"`
	Comment  string          `json:"comment,omitempty"`
}

type HpaOptions struct {
	CpuPercent  int
	MemPercent  int
	NetPercent  int
	DiskPercent int
	MinReplicas int
	MaxReplicas int
}

type HPAState struct {
	Deployed        bool `json:"deployed"`
	Supported       bool `json:"supported"`
	CpuPercernt     int  `json:"cpuPercent"`
	MemoryPercent   int  `json:"memPercent"`
	DiskPercent     int  `json:"diskPercent"`
	NetPercent      int  `json:"netPercent"`
	MinReplicas     int  `json:"minReplicas"`
	MaxReplicas     int  `json:"maxReplicas"`
	CurrentReplicas int  `json:"replicas"`
}

type EventSource struct {
	Component string `json:"component"`
	Host      string `json:"host"`
}

type PodEvent struct {
	FirstTimestamp int64       `json:"firstTimestamp"`
	LastTimestamp  int64       `json:"lastTimestamp"`
	Count          int32       `json:"count"`
	Type           string      `json:"type"`
	Source         EventSource `json:"source"`
	Reason         string      `json:"reason"`
	Message        string      `json:"message"`
}

type PodDesc struct {
	PodDescStr string     `json:"podDescStr"`
	Events     []PodEvent `json:"events"`
}

type JobTemplate struct {
	Name string          `json:"name"`
	Desc json.RawMessage `json:"desc"`
}

type JobState struct {
	Kind                 string   `json:"kind"`
	Reason               string   `json:"reason"`
	JobName              string   `json:"job"`
	State                string   `json:"state"`
	Images               []string `json:"images"`
	PodNum               int      `json:"podnum"`
	Task                 string   `json:"task"`
	User                 string   `json:"user"`
	Workspace            string   `json:"workspace"`
	ClusterIP            string   `json:"clusterip"`
	ParamNum             int      `json:"paramnum"`
	Succeeded            int      `json:"succeeded"`
	Failed               int      `json:"failed"`
	LastScheduledJobName string   `json:"lastScheduledJobName"`
	CreateTime           int64    `json:"createtime"`
	//	LastScheduledTime    string   `json:"lastScheduledTime"`
	//	NextScheduleTime     string   `json:"nextScheduleTime"`
	//	SchedulePeriod       string   `json:"schedulePeriod"`
	LastScheduledTime string `json:"lastScheduledTime"`
	NextScheduleTime  string `json:"nextScheduleTime"`
	SchedulePeriod    string `json:"schedulePeriod"`
}

type JobPodsState struct {
	Job  JobState   `json:job"`
	Pods []PodState `json:"pods"`
}

type TaskState struct {
	TaskName   string     `json:"task"`
	JobNum     int        `json:"jobnum"`
	PodNum     int        `json:"podnum"`
	Jobs       []JobState `json:"jobs"`
	CreateTime int64      `json:"createtime"`
	Group      string     `json:"group"`
	User       string     `json:"user"`
	Workspace  string     `json:"workspace"`
	Kind       string     `json:"kind"`
	State      string     `json:"state"`
	//	LastScheduledTime int64      `json:"lastScheduledTime"`
	//	NextScheduleTime  int64      `json:"nextScheduleTime"`
	//	SchedulePeriod    int64      `json:"schedulePeriod"`
	LastScheduledTime string `json:"lastScheduledTime"`
	NextScheduleTime  string `json:"nextScheduleTime"`
	SchedulePeriod    string `json:"schedulePeriod"`
}

type GroupTasksState struct {
	States    []TaskState `json:"taskstates"`
	GroupName string      `json:"group"`
}
*/
