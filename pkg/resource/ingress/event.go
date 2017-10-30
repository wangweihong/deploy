package ingress

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *IngressManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
