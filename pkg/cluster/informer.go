package cluster

import (
	"fmt"
	"ufleet-deploy/pkg/log"

	"k8s.io/client-go/informers"
	appinformers "k8s.io/client-go/informers/apps/v1beta1"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	batchv2alpa1informers "k8s.io/client-go/informers/batch/v2alpha1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	extensioninformers "k8s.io/client-go/informers/extensions/v1beta1"
	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type ResourceController struct {
	Workspaces map[string]Workspace //这个Workspace引用Cluster的Workspace.

	informerFactory informers.SharedInformerFactory
	//core
	podInformer            coreinformers.PodInformer
	serviceInformer        coreinformers.ServiceInformer
	configmapInformer      coreinformers.ConfigMapInformer
	serviceaccountInformer coreinformers.ServiceAccountInformer
	secretInformer         coreinformers.SecretInformer
	endpointInformer       coreinformers.EndpointsInformer

	//extension
	deploymentInformer extensioninformers.DeploymentInformer
	ingressInformer    extensioninformers.IngressInformer
	daemonsetInformer  extensioninformers.DaemonSetInformer

	//app
	statefulsetInformer appinformers.StatefulSetInformer
	//batch
	cronjobInformer batchv2alpa1informers.CronJobInformer
	jobInformer     batchinformers.JobInformer
}

func (c *ResourceController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh,
		c.podInformer.Informer().HasSynced,
		c.endpointInformer.Informer().HasSynced,
		c.serviceInformer.Informer().HasSynced,
		c.configmapInformer.Informer().HasSynced,
		c.serviceaccountInformer.Informer().HasSynced,
		c.deploymentInformer.Informer().HasSynced,
		c.statefulsetInformer.Informer().HasSynced,
		c.ingressInformer.Informer().HasSynced,
		c.daemonsetInformer.Informer().HasSynced,
		c.cronjobInformer.Informer().HasSynced,
		c.jobInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}

	return nil
}

func ingoreNs(ns string) bool {
	if ns == "kube-system" {
		return true
	}
	return false
}

//所有由ufleet主动创建的资源上都添加特定的标志,表明是ufleet主动创建
//watch到该事件后,无须再创建新的资源对象

//非ufleet创建的资源,只构建资源抽象保存在内存中,不保存在etcd中.(相当于使用了k8s etcd作为缓存)
//在启动的时候,就要开始构建资源抽象.

//只有ufleet创建的资源,才回保存在etcd中.

//在一开始resync时,会触发create事件
//周期性resync时,会触发Update事件
func (c *ResourceController) resourceAdd(obj interface{}) {
	switch r := obj.(type) {
	case *corev1.Pod:
		if ingoreNs(r.Namespace) {
			return
		}

		log.DebugPrint("PodADD %v/%v", r.Namespace, r.Name)

	case *corev1.Service:
		if ingoreNs(r.Namespace) {
			return
		}
	case *corev1.ConfigMap:
		if ingoreNs(r.Namespace) {
			return
		}
	case *corev1.Endpoints:
		if ingoreNs(r.Namespace) {
			return
		}
	case *corev1.ServiceAccount:
		if ingoreNs(r.Namespace) {
			return
		}
	case *corev1.Secret:
		if ingoreNs(r.Namespace) {
			return
		}
	case *extensionsv1beta1.Deployment:
		if ingoreNs(r.Namespace) {
			return
		}
	case *extensionsv1beta1.DaemonSet:
		if ingoreNs(r.Namespace) {
			return
		}
	case *extensionsv1beta1.Ingress:
		if ingoreNs(r.Namespace) {
			return
		}
	case *appv1beta1.StatefulSet:
		if ingoreNs(r.Namespace) {
			return
		}
	case *batchv1.Job:
		if ingoreNs(r.Namespace) {
			return
		}
	case *batchv2alpha1.CronJob:
		if ingoreNs(r.Namespace) {
			return
		}

	}

}

func (c *ResourceController) resourceUpdate(obj, new interface{}) {
	switch obj.(type) {
	case *corev1.Pod:
	case *corev1.Service:
	case *corev1.ConfigMap:
	case *corev1.Endpoints:
	case *corev1.ServiceAccount:
	case *corev1.Secret:
	case *extensionsv1beta1.Deployment:
	case *extensionsv1beta1.DaemonSet:
	case *extensionsv1beta1.Ingress:
	case *appv1beta1.StatefulSet:
	case *batchv1.Job:
	case *batchv2alpha1.CronJob:

	}
}

func (c *ResourceController) resourceDelete(obj interface{}) {
	switch obj.(type) {
	case *corev1.Pod:
	case *corev1.Service:
	case *corev1.ConfigMap:
	case *corev1.Endpoints:
	case *corev1.ServiceAccount:
	case *corev1.Secret:
	case *extensionsv1beta1.Deployment:
	case *extensionsv1beta1.DaemonSet:
	case *extensionsv1beta1.Ingress:
	case *appv1beta1.StatefulSet:
	case *batchv1.Job:
	case *batchv2alpha1.CronJob:

	}
}

func NewResourceController(informerFactory informers.SharedInformerFactory, ws map[string]Workspace) *ResourceController {
	podInformer := informerFactory.Core().V1().Pods()
	serviceInformer := informerFactory.Core().V1().Services()
	configmapInformer := informerFactory.Core().V1().ConfigMaps()
	serviceaccountInformer := informerFactory.Core().V1().ServiceAccounts()
	endpointInformer := informerFactory.Core().V1().Endpoints()
	secretInformer := informerFactory.Core().V1().Secrets()
	deploymentInformer := informerFactory.Extensions().V1beta1().Deployments()
	daemonsetInformer := informerFactory.Extensions().V1beta1().DaemonSets()
	ingressInformer := informerFactory.Extensions().V1beta1().Ingresses()
	statefulsetInformer := informerFactory.Apps().V1beta1().StatefulSets()
	jobInformer := informerFactory.Batch().V1().Jobs()
	cronjobInformer := informerFactory.Batch().V2alpha1().CronJobs()

	c := ResourceController{
		informerFactory:        informerFactory,
		podInformer:            podInformer,
		serviceInformer:        serviceInformer,
		configmapInformer:      configmapInformer,
		endpointInformer:       endpointInformer,
		serviceaccountInformer: serviceaccountInformer,
		secretInformer:         secretInformer,
		deploymentInformer:     deploymentInformer,
		daemonsetInformer:      daemonsetInformer,
		ingressInformer:        ingressInformer,
		statefulsetInformer:    statefulsetInformer,
		jobInformer:            jobInformer,
		cronjobInformer:        cronjobInformer,

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
	return &c
}
