package replicaset

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *ReplicaSetManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
