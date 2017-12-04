package hpa

import (
	"strings"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
)

func HandleEventWatchFromK8sCluster(echan chan cluster.Event, kind string, rm *HpaManager) {
	//可以考虑ufleet创建的configmap直接绑定k8s configmap,
	//在k8s configmap更新的时候再更新绑定的k8s  configmap.
	//在删除的,标记底层资源已经被移除.
	log.DebugPrint(" %v cluster  event handler start !", strings.ToUpper(kind))
	defer log.ErrorPrint("%v cluster  event handler finish !", strings.ToUpper(kind))

	for {
		pe := <-echan

		go func(e cluster.Event) {

			switch e.Action {
			case cluster.ActionDelete:
			case cluster.ActionCreate:
			case cluster.ActionUpdate:
				return
			}
		}(pe)

	}
}
