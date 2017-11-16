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

type ContainerImage struct {
	Container string `json:"container"`
	Image     string `json:"image"`
}
