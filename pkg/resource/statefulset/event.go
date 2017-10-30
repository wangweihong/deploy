package statefulset

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *StatefulSetManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
