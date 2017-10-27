package cronjob

import (
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/resource"
)

const (
	backendKind  = backend.ResourceCronJobs
	resourceKind = "CronJob"
)

func Init() {
	be := backend.NewBackendHandler()

	var err error
	Controller, err = InitCronJobController(be)
	if err != nil {
		panic(err.Error())
	}

	backend.RegisterEventHandler(backendKind, rm)

	go resource.HandleEventWatchFromK8sCluster(cluster.CronJobEventChan, resourceKind, rm)
}
