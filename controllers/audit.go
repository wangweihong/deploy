package controllers

import "fmt"

const (
	operateObjectCluster               = "cluster"
	operateObjectApp                   = "App"
	operateObjectPod                   = "Pod"
	operateObjectService               = "Service"
	operateObjectConfigMap             = "ConfigMap"
	operateObjectReplicationController = "ReplicationController"
	operateObjectSecret                = "Secret"
	operateObjectServiceAccount        = "ServiceAccount"
	operateObjectEndpoint              = "Endpoint"
	operateObjectDeployment            = "Deployment"
	operateObjectReplicaSet            = "ReplicaSet"
	operateObjectDaemonSet             = "DaemonSet"
	operateObjectStatefulSet           = "StatefulSet"
	operateObjectIngress               = "Ingress"
	operateObjectJob                   = "Job"
	operateObjectCronJob               = "CronJob"

	operateTypeCreate        = "create"
	operateTypeUpdate        = "update"
	operateTypeRollback      = "rollback"
	operateTypeDelete        = "delete"
	operateTypeStop          = "stop"
	operateTypeStart         = "start"
	operateTypeScale         = "scale"
	operateTypeAddService    = "add service"
	operateTypeStartHPA      = "start autoscale"
	operateTypePauseOrResume = "pause/resume"

	operateTypeDeleteClusterApp = "deleteClusterObjects"
)

type audit struct {
	object  string
	operate string
}

type ability struct {
	object  string
	operate string
}

var (
	abilityMap = map[string]ability{}

	errTokenInvalid     = fmt.Errorf("invalid token")
	errPermessionDenied = fmt.Errorf("permission denied")
	auditMap            = map[string]audit{
		"NewApp": audit{
			object:  operateObjectApp,
			operate: operateTypeCreate,
		},
		"StopApp": audit{
			object:  operateObjectApp,
			operate: operateTypeStop,
		},
		"StartApp": audit{
			object:  operateObjectApp,
			operate: operateTypeStart,
		},
		"DeleteApp": audit{
			object:  operateObjectApp,
			operate: operateTypeDelete,
		},
		"AddService": audit{
			object:  operateObjectApp,
			operate: operateTypeAddService,
		},
		"AddServiceWithoutWorkspace": audit{
			object:  operateObjectApp,
			operate: operateTypeAddService,
		},

		//Pod
		"CreatePod": audit{
			object:  operateObjectPod,
			operate: operateTypeCreate,
		},
		"UpdatePod": audit{
			object:  operateObjectPod,
			operate: operateTypeUpdate,
		},
		"DeletePod": audit{
			object:  operateObjectPod,
			operate: operateTypeDelete,
		},

		//Service
		"CreateService": audit{
			object:  operateObjectService,
			operate: operateTypeCreate,
		},
		"UpdateService": audit{
			object:  operateObjectService,
			operate: operateTypeUpdate,
		},
		"DeleteService": audit{
			object:  operateObjectService,
			operate: operateTypeDelete,
		},

		//ConfigMap
		"CreateConfigMap": audit{
			object:  operateObjectConfigMap,
			operate: operateTypeCreate,
		},
		"CreateConfigMapCustom": audit{
			object:  operateObjectConfigMap,
			operate: operateTypeCreate,
		},
		"UpdateConfigMap": audit{
			object:  operateObjectConfigMap,
			operate: operateTypeUpdate,
		},
		"UpdateConfigMapCustom": audit{
			object:  operateObjectConfigMap,
			operate: operateTypeUpdate,
		},
		"DeleteConfigMap": audit{
			object:  operateObjectConfigMap,
			operate: operateTypeDelete,
		},

		//ReplicationController
		"CreateReplicationController": audit{
			object:  operateObjectReplicationController,
			operate: operateTypeCreate,
		},
		"UpdateReplicationController": audit{
			object:  operateObjectReplicationController,
			operate: operateTypeUpdate,
		},
		"DeleteReplicationController": audit{
			object:  operateObjectReplicationController,
			operate: operateTypeDelete,
		},

		//Secret
		"CreateSecret": audit{
			object:  operateObjectSecret,
			operate: operateTypeCreate,
		},
		"CreateSecretCustom": audit{
			object:  operateObjectSecret,
			operate: operateTypeCreate,
		},
		"UpdateSecret": audit{
			object:  operateObjectSecret,
			operate: operateTypeUpdate,
		},
		"DeleteSecret": audit{
			object:  operateObjectSecret,
			operate: operateTypeDelete,
		},

		//ServiceAccount
		"CreateServiceAccount": audit{
			object:  operateObjectServiceAccount,
			operate: operateTypeCreate,
		},
		"CreateServiceAccountCustom": audit{
			object:  operateObjectServiceAccount,
			operate: operateTypeCreate,
		},
		"UpdateServiceAccount": audit{
			object:  operateObjectServiceAccount,
			operate: operateTypeUpdate,
		},
		"DeleteServiceAccount": audit{
			object:  operateObjectServiceAccount,
			operate: operateTypeDelete,
		},
		"UpdateServiceAccountCustom": audit{
			object:  operateObjectServiceAccount,
			operate: operateTypeUpdate,
		},

		//Endpoint
		"CreateEndpoint": audit{
			object:  operateObjectEndpoint,
			operate: operateTypeCreate,
		},
		"UpdateEndpoint": audit{
			object:  operateObjectEndpoint,
			operate: operateTypeUpdate,
		},
		"DeleteEndpoint": audit{
			object:  operateObjectEndpoint,
			operate: operateTypeDelete,
		},

		//Deployment
		"CreateDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypeCreate,
		},
		"UpdateDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypeUpdate,
		},
		"DeleteDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypeDelete,
		},
		"StartHPA": audit{
			object:  operateObjectDeployment,
			operate: operateTypeStartHPA,
		},
		"ScaleDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypeScale,
		},
		"ScaleDeploymentIncrement": audit{
			object:  operateObjectDeployment,
			operate: operateTypeScale,
		},
		"RollBackDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypeRollback,
		},
		"RollBackResumeOrPauseDeployment": audit{
			object:  operateObjectDeployment,
			operate: operateTypePauseOrResume,
		},

		//DaemonSet
		"CreateDaemonSet": audit{
			object:  operateObjectDaemonSet,
			operate: operateTypeCreate,
		},
		"UpdateDaemonSet": audit{
			object:  operateObjectDaemonSet,
			operate: operateTypeUpdate,
		},
		"DeleteDaemonSet": audit{
			object:  operateObjectDaemonSet,
			operate: operateTypeDelete,
		},
		"RollBackDaemonSet": audit{
			object:  operateObjectDaemonSet,
			operate: operateTypeRollback,
		},

		//ReplicaSet
		"CreateReplicaSet": audit{
			object:  operateObjectReplicaSet,
			operate: operateTypeCreate,
		},
		"UpdateReplicaSet": audit{
			object:  operateObjectReplicaSet,
			operate: operateTypeUpdate,
		},
		"DeleteReplicaSet": audit{
			object:  operateObjectReplicaSet,
			operate: operateTypeDelete,
		},

		//Ingress
		"CreateIngress": audit{
			object:  operateObjectIngress,
			operate: operateTypeCreate,
		},
		"UpdateIngress": audit{
			object:  operateObjectIngress,
			operate: operateTypeUpdate,
		},
		"DeleteIngress": audit{
			object:  operateObjectIngress,
			operate: operateTypeDelete,
		},

		//Job
		"CreateJob": audit{
			object:  operateObjectJob,
			operate: operateTypeCreate,
		},
		"UpdateJob": audit{
			object:  operateObjectJob,
			operate: operateTypeUpdate,
		},
		"DeleteJob": audit{
			object:  operateObjectJob,
			operate: operateTypeDelete,
		},

		//CronJob
		"CreateCronJob": audit{
			object:  operateObjectCronJob,
			operate: operateTypeCreate,
		},
		"UpdateCronJob": audit{
			object:  operateObjectCronJob,
			operate: operateTypeUpdate,
		},
		"DeleteCronJob": audit{
			object:  operateObjectCronJob,
			operate: operateTypeDelete,
		},
		"PauseOrResumeCronJob": audit{
			object:  operateObjectCronJob,
			operate: operateTypePauseOrResume,
		},
	}
)
