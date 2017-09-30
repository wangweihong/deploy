package pod

/*
import (
	"fmt"
	"time"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/pkg/api/v1"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	resourceDef = "pods"
)

func getGroupVersionResource() schema.GroupVersionResource {
	obj := &v1.Pod{}
	return obj.GroupVersionKind().GroupVersion().WithResource(resourceDef)
}

type PodInformed struct {
	Namespace string
	Name      string
}

type ControllersChanList struct {
	DeleteChan chan Pod
	AddChan    chan Pod
	UpdateChan chan Pod
}

type PodController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	ccl             ControllersChanList
}

func (c *PodControllers) Run(stopChn chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("pod informer Failed to sync")
	}
	return nil
}

func (c *PodController) podAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	//glog.Infof("POD CREATED: %s/%s", pod.Namespace, pod.Name)
	fmt.Printf("POD CREATED: %s/%s\n", pod.Namespace, pod.Name)
	c.ccl.AddChan <- PodInformed{Namespace: pod.Namespace, Name: pod.Name}
}

func (c *PodController) podUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)
	//	glog.Infof(
	fmt.Printf(
		"POD UPDATED. %s/%s %s\n",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
	c.ccl.UpdateChan <- PodInformed{Namespace: pod.Namespace, Name: pod.Name}
}

func (c *PodController) podDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	//	glog.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
	fmt.Printf("POD DELETED: %s/%s\n", pod.Namespace, pod.Name)
	c.ccl.DeletChan <- PodInformed{Namespace: pod.Namespace, Name: pod.Name}
}

// NewPodLoggingController creates a PodLoggingController
func NewPodController(informerFactory informers.SharedInformerFactory) *PodController {
	podInformer := informerFactory.Core().V1().Pods()

	ccl := ControllersChanList{
		DeleteChan: make(chan Pod),
		AddChan:    make(chan Pod),
		UpdateChan: make(chan Pod),
	}

	c := &PodController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
		ccl:             ccl,
	}
	podInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.podAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.podUpdate,
			// Called on resource deletion.
			DeleteFunc: c.podDelete,
		},
	)
	return c
}

func New(config *rest.Config) (chan struct{}, *ControllersChanList, error) {

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	stopCh := make(chan struct{})
	sharedInformerFactory := informers.NewSharedInformerFactory(c, time.Minute*10)
	controller := NewPodController(sharedInformerFactory)
	err = controller.Run(stopCh)
	if err != nil {
		return nil, nil, err
	}
	return stopCh, controller.ccl, nil

}
*/
