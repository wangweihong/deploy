package daemonset

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *DaemonSetManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
