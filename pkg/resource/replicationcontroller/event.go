package replicationcontroller

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *ReplicationControllerManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
