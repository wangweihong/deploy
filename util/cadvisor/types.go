package cadvisor

import "time"

type CpuStat struct {
	Usage float64 `json:"usage"`
}

type NetworkStat struct {
	RxBytes uint64 `json:"rxbytes`
	TxBytes uint64 `json:"txbytes"`
}

type MemoryStat struct {
	Used uint64 `json:"used"`
}

type ContainerStat struct {
	Name    string      `json:"name"`
	Start   time.Time   `json:"start"`
	End     time.Time   `json:"end"`
	Cpu     CpuStat     `json:"cpu"`
	Network NetworkStat `json:"network"`
	Memory  MemoryStat  `json:"memory"`
}
