package cluster

import (
	"fmt"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/sign"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/informers"
	appinformers "k8s.io/client-go/informers/apps/v1beta1"
	autoscalinginformers "k8s.io/client-go/informers/autoscaling/v1"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	batchv2alpa1informers "k8s.io/client-go/informers/batch/v2alpha1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	extensioninformers "k8s.io/client-go/informers/extensions/v1beta1"
	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	autoscalingv1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type ResourceController struct {
	Workspaces map[string]Workspace //这个Workspace引用Cluster的Workspace.

	informerFactory informers.SharedInformerFactory
	//core
	replicationcontrollerInformer coreinformers.ReplicationControllerInformer
	podInformer                   coreinformers.PodInformer
	serviceInformer               coreinformers.ServiceInformer
	configmapInformer             coreinformers.ConfigMapInformer
	serviceaccountInformer        coreinformers.ServiceAccountInformer
	secretInformer                coreinformers.SecretInformer
	endpointInformer              coreinformers.EndpointsInformer

	//extension
	deploymentInformer extensioninformers.DeploymentInformer
	ingressInformer    extensioninformers.IngressInformer
	daemonsetInformer  extensioninformers.DaemonSetInformer
	replicasetInformer extensioninformers.ReplicaSetInformer

	//app
	statefulsetInformer appinformers.StatefulSetInformer
	//batch
	cronjobInformer batchv2alpa1informers.CronJobInformer
	jobInformer     batchinformers.JobInformer
	hpaInformer     autoscalinginformers.HorizontalPodAutoscalerInformer
}

func (c *ResourceController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh,
		c.podInformer.Informer().HasSynced,
		c.endpointInformer.Informer().HasSynced,
		c.serviceInformer.Informer().HasSynced,
		c.secretInformer.Informer().HasSynced,
		c.replicationcontrollerInformer.Informer().HasSynced,
		c.configmapInformer.Informer().HasSynced,
		c.serviceaccountInformer.Informer().HasSynced,
		c.deploymentInformer.Informer().HasSynced,
		c.replicasetInformer.Informer().HasSynced,
		c.statefulsetInformer.Informer().HasSynced,
		c.ingressInformer.Informer().HasSynced,
		c.daemonsetInformer.Informer().HasSynced,
		c.cronjobInformer.Informer().HasSynced,
		c.jobInformer.Informer().HasSynced,
		c.hpaInformer.Informer().HasSynced,
	) {
		return fmt.Errorf("Failed to sync")
	}

	return nil
}

func (c *ResourceController) locateResourceGW(ns string) *Workspace {
	wg, ok := c.Workspaces[ns]
	if !ok {
		return nil
	}
	return &wg
}

func (c *ResourceController) generateEventFromObj(obj interface{}, action ActionType) (*Event, error) {

	//参考自:k8s.io/client-go/tools/cache/index.go的MetaNamespaceIndexFunc()
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	ns := accessor.GetNamespace()
	name := accessor.GetName()
	annotations := accessor.GetAnnotations()
	//由rc创建出来的pod,deployment创建出来的rc包含owner references
	ows := accessor.GetOwnerReferences()

	wg := c.locateResourceGW(ns)
	if wg == nil {
		return nil, nil
	}

	var e Event
	e.Action = action
	e.Group = wg.Group
	e.Workspace = wg.Name
	e.Object = obj
	e.Name = name

	_, ok := annotations[sign.SignFromUfleetKey]
	if ok {
		//		if len(ows) == 0 {
		//Replicaset会继承Deployment的Anntations
		_, ok := obj.(*extensionsv1beta1.ReplicaSet)
		if ok && len(ows) != 0 {
			//这是由deployment创建出来的
			e.FromUfleet = false
		} else {
			e.FromUfleet = true
		}
	}
	return &e, nil
}

//所有由ufleet主动创建的资源上都添加特定的标志,表明是ufleet主动创建
//watch到该事件后,无须再创建新的资源对象

//非ufleet创建的资源,只构建资源抽象保存在内存中,不保存在etcd中.(相当于使用了k8s etcd作为缓存)
//在启动的时候,就要开始构建资源抽象.

//只有ufleet创建的资源,才回保存在etcd中.

//在一开始resync时,会触发create事件
//周期性resync时,会触发Update事件
func (c *ResourceController) resourceAdd(obj interface{}) {
	ep, err := c.generateEventFromObj(obj, ActionCreate)
	if err != nil {
		log.ErrorPrint(err)
		return
	}
	//忽略我们不关心的namespace的资源
	if ep == nil {
		return
	}

	e := *ep
	switch obj.(type) {
	case *corev1.Pod:
		PodEventChan <- e
	case *corev1.Service:
		ServiceEventChan <- e
	case *corev1.ConfigMap:
		ConfigMapEventChan <- e
	case *corev1.ReplicationController:
		ReplicationControllerEventChan <- e
	case *corev1.Endpoints:
		EndpointEventChan <- e
	case *corev1.ServiceAccount:
		ServiceAccountEventChan <- e
	case *corev1.Secret:
		SecretEventChan <- e
	case *extensionsv1beta1.Deployment:
		DeploymentEventChan <- e
	case *extensionsv1beta1.ReplicaSet:
		ReplicaSetEventChan <- e
	case *extensionsv1beta1.DaemonSet:
		DaemonSetEventChan <- e
	case *extensionsv1beta1.Ingress:
		IngressEventChan <- e
	case *appv1beta1.StatefulSet:
		StatefulSetEventChan <- e
	case *batchv1.Job:
		JobEventChan <- e
	case *batchv2alpha1.CronJob:
		CronJobEventChan <- e
	case *autoscalingv1.HorizontalPodAutoscaler:
		HPAEventChan <- e
	}

}

func (c *ResourceController) resourceUpdate(obj, new interface{}) {
	ep, err := c.generateEventFromObj(obj, ActionUpdate)
	if err != nil {
		log.ErrorPrint(err)
		return
	}
	//忽略我们不关心的namespace的资源
	if ep == nil {
		return
	}
	e := *ep

	switch obj.(type) {
	case *corev1.Pod:
		PodEventChan <- e
	case *corev1.Service:
		ServiceEventChan <- e
	case *corev1.ConfigMap:
		ConfigMapEventChan <- e
	case *corev1.ReplicationController:
		ReplicationControllerEventChan <- e
	case *corev1.Endpoints:
		EndpointEventChan <- e
	case *corev1.ServiceAccount:
		ServiceAccountEventChan <- e
	case *corev1.Secret:
		SecretEventChan <- e
	case *extensionsv1beta1.Deployment:
		DeploymentEventChan <- e
	case *extensionsv1beta1.ReplicaSet:
		ReplicaSetEventChan <- e
	case *extensionsv1beta1.DaemonSet:
		DaemonSetEventChan <- e
	case *extensionsv1beta1.Ingress:
		IngressEventChan <- e
	case *appv1beta1.StatefulSet:
		StatefulSetEventChan <- e
	case *batchv1.Job:
		JobEventChan <- e
	case *batchv2alpha1.CronJob:
		CronJobEventChan <- e
	case *autoscalingv1.HorizontalPodAutoscaler:
		HPAEventChan <- e
	}

}

func (c *ResourceController) resourceDelete(obj interface{}) {
	ep, err := c.generateEventFromObj(obj, ActionDelete)
	if err != nil {
		log.ErrorPrint(err)
		return
	}
	//忽略我们不关心的namespace的资源
	if ep == nil {
		return
	}

	e := *ep
	switch obj.(type) {
	case *corev1.Pod:
		PodEventChan <- e
	case *corev1.Service:
		ServiceEventChan <- e
	case *corev1.ConfigMap:
		ConfigMapEventChan <- e
	case *corev1.ReplicationController:
		ReplicationControllerEventChan <- e
	case *corev1.Endpoints:
		EndpointEventChan <- e
	case *corev1.ServiceAccount:
		ServiceAccountEventChan <- e
	case *corev1.Secret:
		SecretEventChan <- e
	case *extensionsv1beta1.Deployment:
		DeploymentEventChan <- e
	case *extensionsv1beta1.ReplicaSet:
		ReplicaSetEventChan <- e
	case *extensionsv1beta1.DaemonSet:
		DaemonSetEventChan <- e
	case *extensionsv1beta1.Ingress:
		IngressEventChan <- e
	case *appv1beta1.StatefulSet:
		StatefulSetEventChan <- e
	case *batchv1.Job:
		JobEventChan <- e
	case *batchv2alpha1.CronJob:
		CronJobEventChan <- e
	case *autoscalingv1.HorizontalPodAutoscaler:
		HPAEventChan <- e
	}

}

func NewResourceController(informerFactory informers.SharedInformerFactory, ws map[string]Workspace) *ResourceController {
	podInformer := informerFactory.Core().V1().Pods()
	serviceInformer := informerFactory.Core().V1().Services()
	configmapInformer := informerFactory.Core().V1().ConfigMaps()
	replicationcontrollerInformer := informerFactory.Core().V1().ReplicationControllers()
	serviceaccountInformer := informerFactory.Core().V1().ServiceAccounts()
	endpointInformer := informerFactory.Core().V1().Endpoints()
	secretInformer := informerFactory.Core().V1().Secrets()
	deploymentInformer := informerFactory.Extensions().V1beta1().Deployments()
	replicasetInformer := informerFactory.Extensions().V1beta1().ReplicaSets()
	daemonsetInformer := informerFactory.Extensions().V1beta1().DaemonSets()
	ingressInformer := informerFactory.Extensions().V1beta1().Ingresses()
	statefulsetInformer := informerFactory.Apps().V1beta1().StatefulSets()
	jobInformer := informerFactory.Batch().V1().Jobs()
	cronjobInformer := informerFactory.Batch().V2alpha1().CronJobs()
	hpaInformer := informerFactory.Autoscaling().V1().HorizontalPodAutoscalers()

	c := ResourceController{
		informerFactory:               informerFactory,
		podInformer:                   podInformer,
		serviceInformer:               serviceInformer,
		configmapInformer:             configmapInformer,
		replicationcontrollerInformer: replicationcontrollerInformer,
		endpointInformer:              endpointInformer,
		serviceaccountInformer:        serviceaccountInformer,
		secretInformer:                secretInformer,
		deploymentInformer:            deploymentInformer,
		replicasetInformer:            replicasetInformer,
		daemonsetInformer:             daemonsetInformer,
		ingressInformer:               ingressInformer,
		statefulsetInformer:           statefulsetInformer,
		jobInformer:                   jobInformer,
		cronjobInformer:               cronjobInformer,
		hpaInformer:                   hpaInformer,

		Workspaces: ws,
	}

	podInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)

	serviceInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	configmapInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	replicationcontrollerInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	serviceaccountInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	secretInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)

	endpointInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)

	deploymentInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	replicasetInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	ingressInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)

	daemonsetInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	statefulsetInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	cronjobInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	jobInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)

	hpaInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.resourceAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.resourceUpdate,
			// Called on resource deletion.
			DeleteFunc: c.resourceDelete,
		},
	)
	return &c
}
