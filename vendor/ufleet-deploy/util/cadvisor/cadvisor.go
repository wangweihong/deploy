package cadvisor

import (
	"fmt"

	client "github.com/google/cadvisor/client/v2"
	//	"github.com/google/cadvisor/info/v1"
	"github.com/google/cadvisor/info/v2"
)

const (
	defaultCollectTime          = 3 * 60 //两分钟
	defaultHostNetworkInterface = "cni0"
)

type resourceClient struct {
	*client.Client
}

type Manager interface {
	GetContainerStats(name string) ([]ContainerStat, error)
}

func NewManager(url string) (Manager, error) {
	c, err := client.NewClient(url)
	if err != nil {
		return nil, err

	}

	_, err = c.MachineInfo()
	if err != nil {
		return nil, err
	}
	return &resourceClient{c}, nil
}

func (rc *resourceClient) GetContainerStats(name string) ([]ContainerStat, error) {

	req := v2.RequestOptions{
		//默认采集距当前时间c分钟内的采样数据.
		//	Start: time.Now().Add(-defaultCollectTime * time.Second),
		IdType: name,
		Count:  10,
	}

	//获取主机的统计
	/*
		hi, err := rc.DockerContainer("/", &req)
		if err != nil {
			return nil, err
		}
	*/

	//获取容器的统计
	//	ci, err := rc.DockerContainer(name, &req)
	ci, err := rc.Stats(name, &req)
	if err != nil {
		return nil, err
	}
	css := make([]ContainerStat, 0)
	for _, v := range ci {
		fmt.Println(v)
		fmt.Println("----------")

		//		var cs ContainerStat

	}
	return css, nil

	/*
		for k := 0; k < len(ci.Stats)-1; k++ {
			var cs ContainerStat
			startStats := ci.Stats[k]
			endStats := ci.Stats[k+1]
			cs.Name = ci.Name
			cs.Start = startStats.Timestamp
			cs.End = endStats.Timestamp

			if ci.Spec.HasCpu {
				cs.Cpu.Usage = float64(endStats.Cpu.Usage.Total-startStats.Cpu.Usage.Total) / float64(cs.End.UnixNano()-cs.Start.UnixNano())
			} else {
				cs.Cpu.Usage = 0.0

			}

			if ci.Spec.HasNetwork {
				cs.Network.RxBytes = endStats.Network.InterfaceStats.RxBytes - startStats.Network.InterfaceStats.RxBytes
				cs.Network.TxBytes = endStats.Network.InterfaceStats.TxBytes - startStats.Network.InterfaceStats.TxBytes
			} else {
				cs.Network.RxBytes = 0
				cs.Network.TxBytes = 0
			}

			if ci.Spec.HasMemory {
				cs.Memory.Used = endStats.Memory.Usage - startStats.Memory.Usage
			} else {
				cs.Memory.Used = 0
			}
			css = append(css, cs)
		}

	*/
	return css, nil
}
