package endpoint

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/resource"
)

func (c *EndpointManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)

}
