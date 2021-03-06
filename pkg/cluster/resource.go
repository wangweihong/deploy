package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
	"ufleet-deploy/pkg/log"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	//	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/labels"

	apiequality "k8s.io/apimachinery/pkg/api/equality"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
	//	"k8s.io/client-go/pkg/api"
	//	corev1 "k8s.io/client-go/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	//	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	appv1beta1 "k8s.io/api/apps/v1beta1"
	appv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/meta"
	watch "k8s.io/apimachinery/pkg/watch"
	kubernetesapi "k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/controller"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
)

var (
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

type GetOptions struct {
	Direct bool
}

//Get从Informer中获取
//Delete/Create则调用k8s相应的接口

/* ----------------- Pod ----------------------*/
type PodHandler interface {
	Get(namespace string, name string, opt GetOptions) (*corev1.Pod, error)
	Delete(namespace string, name string) error
	Create(namespace string, pod *corev1.Pod) error
	Log(namespace, podName string, containerName string, opt LogOption) (string, error)
	Event(namespace, resourceName string) ([]corev1.Event, error)
	Update(namespace string, pod *corev1.Pod) error
	List(namespace string) ([]*corev1.Pod, error)
	GetServices(namespace string, name string) ([]*corev1.Service, error)
	GetCreator(namespace string, name string) (*corev1.SerializedReference, error)
}

func NewPodHandler(group, workspace string) (PodHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &podHandler{Cluster: Cluster}, nil
}

type podHandler struct {
	*Cluster
}

func (h *podHandler) Get(namespace, name string, opt GetOptions) (*corev1.Pod, error) {

	if opt.Direct {
		pod, err := h.clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		//	pod.APIVersion = typemeta.Version
		return pod, nil
	}
	return h.informerController.podInformer.Lister().Pods(namespace).Get(name)
}

func (h *podHandler) Create(namespace string, pod *corev1.Pod) error {
	_, err := h.clientset.CoreV1().Pods(namespace).Create(pod)
	return err
}

func (h *podHandler) Update(namespace string, newpod *corev1.Pod) error {
	_, err := h.clientset.CoreV1().Pods(namespace).Update(newpod)
	return err
}

func (h *podHandler) Delete(namespace, podName string) error {
	return h.clientset.CoreV1().Pods(namespace).Delete(podName, nil)
}

type LogOption struct {
	DisplayTailLine int64
	Timestamps      bool
	SinceSeconds    int64
}

func (h *podHandler) Log(namespace, podName string, containerName string, opt LogOption) (string, error) {
	corev1Opt := corev1.PodLogOptions{
		Container:    containerName,
		TailLines:    &opt.DisplayTailLine,
		Timestamps:   opt.Timestamps,
		SinceSeconds: &opt.SinceSeconds,
	}

	req := h.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1Opt)
	bc, err := req.Do().Raw()
	if err != nil {
		return "", err
	}

	return string(bc), nil
}

func (h *podHandler) Event(namespace, podName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&podName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

func (h *podHandler) List(namespace string) ([]*corev1.Pod, error) {
	return h.informerController.podInformer.Lister().Pods(namespace).List(labels.Everything())
}

func (h *podHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	pod, err := h.informerController.podInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(pod.Labels)) {
			services = append(services, service)
		}
	}
	return services, nil
}

func (h *podHandler) GetCreator(namespace string, name string) (*corev1.SerializedReference, error) {
	pod, err := h.informerController.podInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	sr, err := getResourceCreator(pod)
	if err != nil {
		return nil, err
	}

	if sr == nil {
		return nil, nil
	}
	v1sr := kubernetesapiSerrializedReferenceToClientGo(*sr)

	return &v1sr, nil
}

/* ----------------- Service ----------------------*/

type ServiceHandler interface {
	Get(namespace string, name string) (*corev1.Service, error)
	Delete(namespace string, name string) error
	Create(namespace string, service *corev1.Service) error
	Update(namespace string, service *corev1.Service) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	List(namespace string) ([]*corev1.Service, error)
	GetReferenceResources(namespace, name string) ([]corev1.ObjectReference, error)
	GetIngresses(namespace, name string) ([]*extensionsv1beta1.Ingress, error)
}

func NewServiceHandler(group, workspace string) (ServiceHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &serviceHandler{Cluster: Cluster}, nil
}

type serviceHandler struct {
	*Cluster
}

func (h *serviceHandler) Get(namespace, name string) (*corev1.Service, error) {
	return h.informerController.serviceInformer.Lister().Services(namespace).Get(name)
}

func (h *serviceHandler) Create(namespace string, service *corev1.Service) error {
	_, err := h.clientset.CoreV1().Services(namespace).Create(service)
	return err
}

func (h *serviceHandler) Update(namespace string, service *corev1.Service) error {
	_, err := h.clientset.CoreV1().Services(namespace).Update(service)
	return err
}

func (h *serviceHandler) Delete(namespace, serviceName string) error {
	return h.clientset.CoreV1().Services(namespace).Delete(serviceName, nil)
}

func (h *serviceHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	svc, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	selectorStr := svc.Spec.Selector
	selector := labels.Set(selectorStr).AsSelector()
	pods, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (h *serviceHandler) List(namespace string) ([]*corev1.Service, error) {
	return h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
}

func (h *serviceHandler) GetIngresses(namespace, name string) ([]*extensionsv1beta1.Ingress, error) {
	_, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	allings, err := h.informerController.ingressInformer.Lister().Ingresses(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ings := make([]*extensionsv1beta1.Ingress, 0)
	for k := range allings {
		ing := allings[k]
		if ing.Spec.Backend != nil {
			if ing.Spec.Backend.ServiceName == name {
				ings = append(ings, ing)
				continue
			}
		}

		for _, v := range ing.Spec.Rules {
			if v.HTTP != nil {
				for _, j := range v.HTTP.Paths {
					if j.Backend.ServiceName == name {
						ings = append(ings, ing)
						goto out
					}
				}
			}
		}
	out:
	}
	return ings, nil
}

func (h *serviceHandler) GetReferenceResources(namespace, name string) ([]corev1.ObjectReference, error) {
	s, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	ors := make([]corev1.ObjectReference, 0)

	if s.Spec.Selector == nil {
		return ors, nil
	}

	selector := labels.Set(s.Spec.Selector).AsSelectorPreValidated()
	slabels := s.Spec.Selector
	log.DebugPrint("service'labels:", slabels)

	//statefulset
	allsfs, err := h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	//daemonset,需要使用的是Template的
	alldts, err := h.informerController.daemonsetInformer.Lister().DaemonSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	//deployment
	allds, err := h.informerController.deploymentInformer.Lister().Deployments(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	//replicaset
	allrss, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	//replicationcontroller
	allrcs, err := h.informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	//pod
	allps, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	requirements := make([]labels.Requirement, 0)
	for l, v := range slabels {
		newreq, err := labels.NewRequirement(l, selection.In, []string{v})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, *newreq)
	}
	inSelector := labels.NewSelector()
	inSelector = inSelector.Add(requirements...)
	log.DebugPrint("inSelector:", inSelector.String())

	//需要将检测service labels是否是statefulset的pod template的子集

	sfs := make([]*appv1beta2.StatefulSet, 0)
	for _, v := range allsfs {
		if inSelector.Matches(labels.Set(v.Spec.Template.Labels)) {
			log.DebugPrint("statefulset: %v: matchlabels:%v", v.Name, v.Spec.Template.Labels)
			sfs = append(sfs, v)
		}
	}

	dts := make([]*extensionsv1beta1.DaemonSet, 0)
	for _, v := range alldts {
		if inSelector.Matches(labels.Set(v.Spec.Template.Labels)) {
			log.DebugPrint("daemonset %v: matchlabels:%v", v.Name, v.Spec.Template.Labels)
			dts = append(dts, v)
		}
	}

	ds := make([]*extensionsv1beta1.Deployment, 0)
	for _, v := range allds {
		if inSelector.Matches(labels.Set(v.Spec.Template.Labels)) {
			log.DebugPrint("deployment: %v: matchlabels:%v", v.Name, v.Spec.Template.Labels)
			ds = append(ds, v)
		}
	}

	rss := make([]*extensionsv1beta1.ReplicaSet, 0)
	for _, v := range allrss {
		if inSelector.Matches(labels.Set(v.Spec.Template.Labels)) {
			log.DebugPrint("replicaset: %v: matchlabels:%v", v.Name, v.Spec.Template.Labels)
			rss = append(rss, v)
		}
	}

	rcs := make([]*corev1.ReplicationController, 0)
	for _, v := range allrcs {
		if inSelector.Matches(labels.Set(v.Spec.Selector)) {
			log.DebugPrint("replicationcontroller: %v: matchlabels:%v", v.Name, v.Spec.Template.Labels)
			rcs = append(rcs, v)
		}
	}

	ps := make([]*corev1.Pod, 0)
	for _, v := range allps {
		if v.Labels != nil {
			if inSelector.Matches(labels.Set(v.Labels)) {
				log.DebugPrint("replicaset: %v: matchlabels:%v", v.Name, v.Labels)
				ps = append(ps, v)
			}
		}
	}

	ors = append(ors, runtimeObjectListToObjectReference(sfs)...)
	ors = append(ors, runtimeObjectListToObjectReference(dts)...)
	ors = append(ors, runtimeObjectListToObjectReference(ds)...)
	ors = append(ors, runtimeObjectListToObjectReference(rss)...)
	ors = append(ors, runtimeObjectListToObjectReference(rcs)...)
	ors = append(ors, runtimeObjectListToObjectReference(ps)...)

	return ors, nil
}

/* ------------------------ Configmap ----------------------------*/

type ConfigMapHandler interface {
	Get(namespace string, name string) (*corev1.ConfigMap, error)
	Create(namespace string, cm *corev1.ConfigMap) error
	Delete(namespace string, name string) error
	Update(namespace string, service *corev1.ConfigMap) error
	List(namespace string) ([]*corev1.ConfigMap, error)
	GetReferenceResources(namespace string, name string) ([]corev1.ObjectReference, error)
}

func NewConfigMapHandler(group, workspace string) (ConfigMapHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &configmapHandler{Cluster: Cluster}, nil
}

type configmapHandler struct {
	*Cluster
}

func (h *configmapHandler) Get(namespace, name string) (*corev1.ConfigMap, error) {
	return h.informerController.configmapInformer.Lister().ConfigMaps(namespace).Get(name)
}

func (h *configmapHandler) Create(namespace string, configmap *corev1.ConfigMap) error {
	_, err := h.clientset.CoreV1().ConfigMaps(namespace).Create(configmap)
	return err
}

func (h *configmapHandler) Update(namespace string, resource *corev1.ConfigMap) error {
	_, err := h.clientset.CoreV1().ConfigMaps(namespace).Update(resource)
	return err
}

func (h *configmapHandler) Delete(namespace, configmapName string) error {
	return h.clientset.CoreV1().ConfigMaps(namespace).Delete(configmapName, nil)
}

func (h *configmapHandler) List(namespace string) ([]*corev1.ConfigMap, error) {
	return h.informerController.configmapInformer.Lister().ConfigMaps(namespace).List(labels.Everything())
}

func (h *configmapHandler) GetReferenceResources(namespace string, name string) ([]corev1.ObjectReference, error) {
	_, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	ors, err := getGeneralResourceReference(h.informerController, namespace, name, IsPodSpecReferenceConfigMap)
	if err != nil {
		return nil, err
	}
	return ors, nil

}

/* ------------------------ ReplicationController---------------------------*/

type ReplicationControllerHandler interface {
	Get(namespace string, name string) (*corev1.ReplicationController, error)
	Create(namespace string, cm *corev1.ReplicationController) error
	Delete(namespace string, name string) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	Update(namespace string, resource *corev1.ReplicationController) error
	Scale(namespace, name string, num int32) error
	Event(namespace, resourceName string) ([]corev1.Event, error)
	List(namespace string) ([]*corev1.ReplicationController, error)
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewReplicationControllerHandler(group, workspace string) (ReplicationControllerHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &replicationcontrollerHandler{Cluster: Cluster}, nil
}

type replicationcontrollerHandler struct {
	*Cluster
}

func (h *replicationcontrollerHandler) Get(namespace, name string) (*corev1.ReplicationController, error) {
	return h.informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).Get(name)
}

func (h *replicationcontrollerHandler) Create(namespace string, replicationcontroller *corev1.ReplicationController) error {
	_, err := h.clientset.CoreV1().ReplicationControllers(namespace).Create(replicationcontroller)
	return err
}

func (h *replicationcontrollerHandler) Update(namespace string, resource *corev1.ReplicationController) error {
	_, err := h.clientset.CoreV1().ReplicationControllers(namespace).Update(resource)
	return err
}

func (h *replicationcontrollerHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

func (h *replicationcontrollerHandler) Delete(namespace, replicationcontrollerName string) error {
	return h.clientset.CoreV1().ReplicationControllers(namespace).Delete(replicationcontrollerName, nil)
}
func (h *replicationcontrollerHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	d, err := h.informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).Get(name)
	if err != nil {
		return nil, nil
	}
	//rsSelector := d.Spec.Selector.MatchLabels
	rsSelector := d.Spec.Selector

	selector := labels.Set(rsSelector).AsSelector()
	//opts := corev1.ListOptions{LabelSelector: selector.String()}
	//po, err := h.clientset.CoreV1().Pods(namespace).List(opts)
	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if controllerRef.UID == d.UID {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}
func (h *replicationcontrollerHandler) Scale(namespace, replicationcontrollerName string, num int32) error {
	rc, err := h.Get(namespace, replicationcontrollerName)
	if err != nil {
		return err
	}

	if rc.Spec.Replicas != nil {
		if *rc.Spec.Replicas == num {
			return nil
		}
	} else {
		if num == 1 {
			return nil
		}
	}

	rc.Spec.Replicas = &num
	oldrv := rc.ResourceVersion
	rc.ResourceVersion = ""
	err = h.Update(namespace, rc)
	if err != nil {
		return err
	}

	for {
		newrc, err := h.Get(namespace, replicationcontrollerName)
		if err != nil {
			return err
		}
		if newrc.ResourceVersion == oldrv {
			time.Sleep(500 * time.Microsecond)
			continue
		}

		if *newrc.Spec.Replicas > newrc.Status.Replicas {
			for _, v := range newrc.Status.Conditions {
				if v.Type == corev1.ReplicationControllerReplicaFailure {
					if v.Status == corev1.ConditionTrue {
						return fmt.Errorf(v.Message)
					}
				}
			}
		} else {
			return nil
		}
	}
	return nil
}

func (h *replicationcontrollerHandler) List(namespace string) ([]*corev1.ReplicationController, error) {
	return h.informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).List(labels.Everything())
}

func (h *replicationcontrollerHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	d, err := h.informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(d.Spec.Selector)) {
			services = append(services, service)
		}
	}
	return services, nil
}

/* ------------------------ ServiceAccount ----------------------------*/

type ServiceAccountHandler interface {
	Get(namespace string, name string) (*corev1.ServiceAccount, error)
	Create(namespace string, sa *corev1.ServiceAccount) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *corev1.ServiceAccount) error
	List(namespace string) ([]*corev1.ServiceAccount, error)
	GetReferenceResources(namespace string, name string) ([]corev1.ObjectReference, error)
}

func NewServiceAccountHandler(group, workspace string) (ServiceAccountHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &serviceaccountHandler{Cluster: Cluster}, nil
}

type serviceaccountHandler struct {
	*Cluster
}

func (h *serviceaccountHandler) Get(namespace, name string) (*corev1.ServiceAccount, error) {
	return h.informerController.serviceaccountInformer.Lister().ServiceAccounts(namespace).Get(name)
}

func (h *serviceaccountHandler) Create(namespace string, serviceaccount *corev1.ServiceAccount) error {
	_, err := h.clientset.CoreV1().ServiceAccounts(namespace).Create(serviceaccount)
	return err
}

func (h *serviceaccountHandler) Update(namespace string, resource *corev1.ServiceAccount) error {
	_, err := h.clientset.CoreV1().ServiceAccounts(namespace).Update(resource)
	return err
}

func (h *serviceaccountHandler) Delete(namespace, serviceaccountName string) error {
	return h.clientset.CoreV1().ServiceAccounts(namespace).Delete(serviceaccountName, nil)
}

func (h *serviceaccountHandler) List(namespace string) ([]*corev1.ServiceAccount, error) {
	return h.informerController.serviceaccountInformer.Lister().ServiceAccounts(namespace).List(labels.Everything())
}

func (h *serviceaccountHandler) GetReferenceResources(namespace string, name string) ([]corev1.ObjectReference, error) {
	_, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	ors, err := getGeneralResourceReference(h.informerController, namespace, name, IsPodSpecReferenceServiceAccount)
	if err != nil {
		return nil, err
	}
	return ors, nil

}

/* ------------------------ Secret ----------------------------*/

type SecretHandler interface {
	Get(namespace string, name string) (*corev1.Secret, error)
	Create(namespace string, secret *corev1.Secret) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *corev1.Secret) error
	List(namespace string) ([]*corev1.Secret, error)
	GetReferenceResources(namespace, name string) ([]corev1.ObjectReference, error)
}

func NewSecretHandler(group, workspace string) (SecretHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &secretHandler{Cluster: Cluster}, nil
}

type secretHandler struct {
	*Cluster
}

func (h *secretHandler) Get(namespace, name string) (*corev1.Secret, error) {
	return h.informerController.secretInformer.Lister().Secrets(namespace).Get(name)
}

func (h *secretHandler) Create(namespace string, secret *corev1.Secret) error {
	_, err := h.clientset.CoreV1().Secrets(namespace).Create(secret)
	return err
}

func (h *secretHandler) Delete(namespace, secretName string) error {
	return h.clientset.CoreV1().Secrets(namespace).Delete(secretName, nil)
}

func (h *secretHandler) Update(namespace string, resource *corev1.Secret) error {
	_, err := h.clientset.CoreV1().Secrets(namespace).Update(resource)
	return err
}

func (h *secretHandler) List(namespace string) ([]*corev1.Secret, error) {
	return h.informerController.secretInformer.Lister().Secrets(namespace).List(labels.Everything())
}

func (h *secretHandler) GetReferenceResources(namespace string, name string) ([]corev1.ObjectReference, error) {
	_, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	ors, err := getGeneralResourceReference(h.informerController, namespace, name, IsPodSpecReferenceSecret)
	if err != nil {
		return nil, err
	}
	return ors, nil

}

/* ------------------------ Endpoint ----------------------------*/

type EndpointHandler interface {
	Get(namespace string, name string) (*corev1.Endpoints, error)
	Create(namespace string, ep *corev1.Endpoints) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *corev1.Endpoints) error
	List(namespace string) ([]*corev1.Endpoints, error)
}

func NewEndpointHandler(group, workspace string) (EndpointHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &endpointHandler{Cluster: Cluster}, nil
}

type endpointHandler struct {
	*Cluster
}

func (h *endpointHandler) Get(namespace, name string) (*corev1.Endpoints, error) {
	return h.informerController.endpointInformer.Lister().Endpoints(namespace).Get(name)
}

func (h *endpointHandler) Create(namespace string, endpoint *corev1.Endpoints) error {
	_, err := h.clientset.CoreV1().Endpoints(namespace).Create(endpoint)
	return err
}

func (h *endpointHandler) Delete(namespace, endpointName string) error {
	return h.clientset.CoreV1().Endpoints(namespace).Delete(endpointName, nil)
}

func (h *endpointHandler) Update(namespace string, resource *corev1.Endpoints) error {
	_, err := h.clientset.CoreV1().Endpoints(namespace).Update(resource)
	return err
}
func (h *endpointHandler) List(namespace string) ([]*corev1.Endpoints, error) {
	return h.informerController.endpointInformer.Lister().Endpoints(namespace).List(labels.Everything())
}

/* ------------------------ Deployment ----------------------------*/

type DeploymentHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.Deployment, error)
	List(namespace string) ([]*extensionsv1beta1.Deployment, error)
	Create(namespace string, d *extensionsv1beta1.Deployment) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *extensionsv1beta1.Deployment) error
	Scale(namespace, name string, num int32) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	Event(namespace, resourceName string) ([]corev1.Event, error)
	Revision(namespace, name string) (*int64, error)
	GetRevisionsAndDescribe(namespace, name string) (map[int64]*corev1.PodTemplateSpec, error)
	GetRevisionsAndReplicas(namespace, name string) (map[int64]*extensionsv1beta1.ReplicaSet, error)
	GetCurrentRevisionAndReplicaSet(namespace, name string) (*int64, *extensionsv1beta1.ReplicaSet, error)
	Rollback(namespace, name string, revision int64) (*string, error)
	ResumeRollout(namespace, name string) error
	PauseRollout(namespace, name string) error
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewDeploymentHandler(group, workspace string) (DeploymentHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &deploymentHandler{Cluster: Cluster}, nil
}

type deploymentHandler struct {
	*Cluster
}

func (h *deploymentHandler) Get(namespace, name string) (*extensionsv1beta1.Deployment, error) {
	return h.informerController.deploymentInformer.Lister().Deployments(namespace).Get(name)
}
func (h *deploymentHandler) List(namespace string) ([]*extensionsv1beta1.Deployment, error) {
	return h.informerController.deploymentInformer.Lister().Deployments(namespace).List(labels.Everything())
}

func (h *deploymentHandler) Create(namespace string, deployment *extensionsv1beta1.Deployment) error {
	_, err := h.clientset.ExtensionsV1beta1().Deployments(namespace).Create(deployment)
	return err
}

func (h *deploymentHandler) Update(namespace string, resource *extensionsv1beta1.Deployment) error {
	_, err := h.clientset.ExtensionsV1beta1().Deployments(namespace).Update(resource)
	return err
}

func (h *deploymentHandler) Delete(namespace, deploymentName string) error {
	//	return h.clientset.ExtensionsV1beta1().Deployments(namespace).Delete(deploymentName, nil)
	var e error
	allOldRSs, newRs, rsErr := h.GetAllReplicaSets(namespace, deploymentName)
	allRSs := allOldRSs
	if newRs != nil {
		allRSs = append(allRSs, newRs)
	}

	allpods, podErr := h.GetPods(namespace, deploymentName)

	err := h.clientset.ExtensionsV1beta1().Deployments(namespace).Delete(deploymentName, nil)
	if err != nil {
		return err
	}

	if rsErr == nil {
		for _, v := range allRSs {
			err := h.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Delete(v.Name, nil)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					e = log.ErrorPrint(fmt.Sprintf("try to delete rs %v fail for %v", v.Name, err))
				}
			}
		}

		if podErr == nil {

			if allpods != nil {
				for _, v := range allpods {
					err := h.clientset.CoreV1().Pods(namespace).Delete(v.Name, nil)
					if err != nil {
						if !apierrors.IsNotFound(err) {
							e = log.ErrorPrint("try to delete po %v fail for %v", v.Name, err)
						}
					}
				}
			}
		}
	}

	if e != nil || rsErr != nil || podErr != nil {
		return fmt.Errorf("delete deployment success, but resources owned by this deployment still exists, please delete them manually")
	}
	return nil
}

func (h *deploymentHandler) Scale(namespace, name string, num int32) error {
	d, err := h.informerController.deploymentInformer.Lister().Deployments(namespace).Get(name)
	if err != nil {
		return err
	}

	d.Spec.Replicas = &num
	d.ResourceVersion = ""
	_, err = h.clientset.ExtensionsV1beta1().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}

	//扩容时,可能出现资源不足,导致创建失败;检测如果因为资源不足创建失败,则报错
	time.Sleep(500 * time.Microsecond)
	for {
		//
		//		d, err := h.clientset.ExtensionsV1beta1().Deployments(namespace).Get(name, meta_v1.GetOptions{})
		d, err := h.informerController.deploymentInformer.Lister().Deployments(namespace).Get(name)
		if err != nil {
			return err
		}
		if *d.Spec.Replicas > d.Status.Replicas {
			for _, v := range d.Status.Conditions {
				if v.Type == extensionsv1beta1.DeploymentReplicaFailure {
					if v.Status == corev1.ConditionTrue {
						return fmt.Errorf(v.Message)
					}
				}
			}
		} else {
			return nil
		}
	}
}

func (h *deploymentHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	d, err := h.informerController.deploymentInformer.Lister().Deployments(namespace).Get(name)
	if err != nil {
		return nil, nil
	}
	//	rsSelector := d.Spec.Selector.MatchLabels
	rsSelector := d.Spec.Template.Labels
	selector := labels.Set(rsSelector).AsSelector()
	//opts := corev1.ListOptions{LabelSelector: selector.String()}
	//po, err := h.clientset.CoreV1().Pods(namespace).List(opts)
	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	//当deployment修改podtemplate selectors时, deployment将会和之前的Replicaset脱离Owner关系
	//deployment将会创建的replicaset,同时生成新的Pod.对于已经脱离Owner的Replicaset,除非
	//deployment修改回之前的selectors,否则互不影响.
	allOldRSs, newRs, err := h.GetAllReplicaSets(namespace, name)
	if err != nil {
		return nil, err
	}

	rsList := allOldRSs
	if newRs != nil {
		rsList = append(rsList, newRs)
	}

	podMap := make(map[types.UID]struct{}, len(rsList))
	for _, v := range rsList {
		podMap[v.UID] = struct{}{}
	}

	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if _, ok := podMap[controllerRef.UID]; ok {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}

func (h *deploymentHandler) Rollback(namespace, name string, revision int64) (*string, error) {
	d, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	if d.Spec.Paused {
		return nil, fmt.Errorf("deployment is paused, cannot rollback")
	}

	rm, err := h.GetRevisionsAndDescribe(namespace, name)
	if err != nil {
		return nil, err
	}

	if len(rm) == 0 {
		s := fmt.Sprintf("not rollout history found")
		return &s, nil
	}

	_, ok := rm[revision]
	if !ok {
		return nil, fmt.Errorf("revision is not found")
	}

	drb := &extensionsv1beta1.DeploymentRollback{
		Name: name,
		RollbackTo: extensionsv1beta1.RollbackConfig{
			Revision: revision,
		},
	}

	event, err := h.clientset.CoreV1().Events(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	err = h.clientset.ExtensionsV1beta1().Deployments(namespace).Rollback(drb)
	if err != nil {
		return nil, err
	}

	watch, err := h.clientset.CoreV1().Events(namespace).Watch(metav1.ListOptions{Watch: true, ResourceVersion: event.ResourceVersion})
	if err != nil {
		return nil, err
	}

	result := ""
	result = watchRollbackEvent(watch)

	return &result, nil
}

func (h *deploymentHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}
func (h *deploymentHandler) ResumeRollout(namespace, name string) error {
	d, err := h.Get(namespace, name)
	if err != nil {
		return err
	}

	if !d.Spec.Paused {
		return fmt.Errorf("deployments \"%v\" is not paused")
	}
	d.Spec.Paused = false
	_, err = h.clientset.ExtensionsV1beta1().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}
	return nil

}

func (h *deploymentHandler) PauseRollout(namespace, name string) error {
	d, err := h.Get(namespace, name)
	if err != nil {
		return err
	}

	if d.Spec.Paused {
		return fmt.Errorf("deployments \"%v\" is already paused")
	}
	d.Spec.Paused = true
	_, err = h.clientset.ExtensionsV1beta1().Deployments(namespace).Update(d)
	if err != nil {
		return err
	}
	return nil
}
func (h *deploymentHandler) Revision(namespace, name string) (*int64, error) {
	d, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	revision, err := GetCurrentDeploymentRevision(d)
	if err != nil {
		return nil, err
	}

	return &revision, nil
}

func EqualIgnoreHash(template1, template2 *corev1.PodTemplateSpec) (bool, error) {
	t1Copy := template1.DeepCopy() //api.Scheme.DeepCopy(template1)
	//t1Copy := cp.(*corev1.PodTemplateSpec)
	t2Copy := template2.DeepCopy()
	//	t2Copy := cp.(*corev1.PodTemplateSpec)
	// First, compare template.Labels (ignoring hash)
	labels1, labels2 := t1Copy.Labels, t2Copy.Labels
	if len(labels1) > len(labels2) {
		labels1, labels2 = labels2, labels1
	}
	// We make sure len(labels2) >= len(labels1)
	for k, v := range labels2 {
		if labels1[k] != v && k != extensionsv1beta1.DefaultDeploymentUniqueLabelKey {
			return false, nil
		}
	}
	// Then, compare the templates without comparing their labels
	t1Copy.Labels, t2Copy.Labels = nil, nil
	return apiequality.Semantic.DeepEqual(t1Copy, t2Copy), nil
}

func (h *deploymentHandler) GetAllReplicaSets(namespace string, name string) ([]*extensionsv1beta1.ReplicaSet, *extensionsv1beta1.ReplicaSet, error) {
	d, err := h.Get(namespace, name)
	if err != nil {
		return nil, nil, err
	}
	deploymentSelector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("deployment %s/%s has invalid label selector: %v", d.Namespace, d.Name, err)
	}

	rsList, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).List(deploymentSelector)
	if err != nil {
		return nil, nil, err
	}

	owned := make([]*extensionsv1beta1.ReplicaSet, 0)

	for _, v := range rsList {
		controllerRef := metav1.GetControllerOf(v)
		if controllerRef != nil && controllerRef.UID == d.UID {
			owned = append(owned, v)
		}
	}

	var newRS *extensionsv1beta1.ReplicaSet
	for i := range owned {
		equal, err := EqualIgnoreHash(&rsList[i].Spec.Template, &d.Spec.Template)
		if err != nil {
			return nil, nil, err
		}
		if equal {
			// In rare cases, such as after cluster upgrades, Deployment may end up with
			// having more than one new ReplicaSets that have the same template as its template,
			// see https://github.com/kubernetes/kubernetes/issues/40415
			// We deterministically choose the oldest new ReplicaSet.
			newRS = owned[i]
			break
		}
	}
	var allOldRSs []*extensionsv1beta1.ReplicaSet

	for _, rs := range owned {
		// Filter out new replica set
		//过滤掉当前deployment的RS
		if newRS != nil && rs.UID == newRS.UID {
			continue
		}
		allOldRSs = append(allOldRSs, rs)
	}

	// new ReplicaSet does not exist.
	return allOldRSs, newRS, nil

}

func GetCurrentDeploymentRevision(d *extensionsv1beta1.Deployment) (int64, error) {
	revisionStr, ok := d.Annotations[deploymentutil.RevisionAnnotation]
	if !ok {
		return 0, fmt.Errorf("revision doesn't exists")
	}

	revision, err := strconv.ParseInt(revisionStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return revision, nil
}

func (h *deploymentHandler) GetRevisionsAndDescribe(namespace, name string) (map[int64]*corev1.PodTemplateSpec, error) {
	allOldRSs, newRs, err := h.GetAllReplicaSets(namespace, name)
	if err != nil {
		return nil, err
	}

	allRSs := allOldRSs
	if newRs != nil {
		allRSs = append(allRSs, newRs)
	}

	revisionToSpec := make(map[int64]*corev1.PodTemplateSpec)
	for _, rs := range allRSs {
		v, err := deploymentutil.Revision(rs)
		if err != nil {
			log.ErrorPrint(err)
			continue
		}
		revisionToSpec[v] = &rs.Spec.Template
	}
	return revisionToSpec, nil
}

func (h *deploymentHandler) GetCurrentRevisionAndReplicaSet(namespace, name string) (*int64, *extensionsv1beta1.ReplicaSet, error) {
	d, err := h.Get(namespace, name)
	if err != nil {
		return nil, nil, err
	}
	rev, err := GetCurrentDeploymentRevision(d)
	if err != nil {
		return nil, nil, err
	}

	_, newRs, err := h.GetAllReplicaSets(namespace, name)
	if err != nil {
		return nil, nil, err
	}

	/*
		allRSs := allOldRSs
		if newRs != nil {
			allRSs = append(allRSs, newRs)
		}

		revisionToReplicas := make(map[int64]*extensionsv1beta1.ReplicaSet)
		for _, rs := range allRSs {
			v, err := deploymentutil.Revision(rs)
			if err != nil {
				log.DebugPrint(err)
				continue
			}
			revisionToReplicas[v] = rs
		}
		rs, ok := revisionToReplicas[rev]
		if !ok {
			return nil, nil, nil
		}
	*/
	if newRs != nil {
		return &rev, newRs, nil
	}

	return nil, nil, nil
}

func (h *deploymentHandler) GetRevisionsAndReplicas(namespace, name string) (map[int64]*extensionsv1beta1.ReplicaSet, error) {
	allOldRSs, newRs, err := h.GetAllReplicaSets(namespace, name)
	if err != nil {
		return nil, err
	}

	allRSs := allOldRSs
	if newRs != nil {
		allRSs = append(allRSs, newRs)
	}

	revisionToReplicas := make(map[int64]*extensionsv1beta1.ReplicaSet)
	for _, rs := range allRSs {
		v, err := deploymentutil.Revision(rs)
		if err != nil {
			log.ErrorPrint(err)
			continue
		}
		revisionToReplicas[v] = rs
	}
	return revisionToReplicas, nil
}

func (h *deploymentHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	deploy, err := h.informerController.deploymentInformer.Lister().Deployments(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(deploy.Spec.Template.Labels)) {
			services = append(services, service)
		}
	}
	return services, nil
}

/* ------------------------ ReplicaSet---------------------------*/

type ReplicaSetHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.ReplicaSet, error)
	List(namespace string) ([]*extensionsv1beta1.ReplicaSet, error)
	Create(namespace string, cm *extensionsv1beta1.ReplicaSet) error
	Delete(namespace string, name string) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	Update(namespace string, resource *extensionsv1beta1.ReplicaSet) error
	Scale(namespace, name string, num int32) error
	Event(namespace, resourceName string) ([]corev1.Event, error)
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewReplicaSetHandler(group, workspace string) (ReplicaSetHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &replicasetHandler{Cluster: Cluster}, nil
}

type replicasetHandler struct {
	*Cluster
}

func (h *replicasetHandler) Get(namespace, name string) (*extensionsv1beta1.ReplicaSet, error) {
	return h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).Get(name)
}
func (h *replicasetHandler) List(namespace string) ([]*extensionsv1beta1.ReplicaSet, error) {
	return h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).List(labels.Everything())
}

func (h *replicasetHandler) Create(namespace string, replicaset *extensionsv1beta1.ReplicaSet) error {
	_, err := h.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Create(replicaset)
	return err
}

func (h *replicasetHandler) Update(namespace string, resource *extensionsv1beta1.ReplicaSet) error {
	_, err := h.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Update(resource)
	return err
}

func (h *replicasetHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

func (h *replicasetHandler) Delete(namespace, replicasetName string) error {
	return h.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Delete(replicasetName, nil)
}
func (h *replicasetHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	d, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).Get(name)
	if err != nil {
		return nil, nil
	}
	//rsSelector := d.Spec.Selector.MatchLabels
	rsSelector := d.Spec.Selector.MatchLabels
	selector := labels.Set(rsSelector).AsSelector()

	//	selector := labels.Set(rsSelector).AsSelector()
	//opts := corev1.ListOptions{LabelSelector: selector.String()}
	//po, err := h.clientset.ExtensionsV1beta1().Pods(namespace).List(opts)
	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if controllerRef.UID == d.UID {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}
func (h *replicasetHandler) Scale(namespace, replicasetName string, num int32) error {
	rc, err := h.Get(namespace, replicasetName)
	if err != nil {
		return err
	}

	if rc.Spec.Replicas != nil {
		if *rc.Spec.Replicas == num {
			return nil
		}
	} else {
		if num == 1 {
			return nil
		}
	}

	rc.Spec.Replicas = &num
	oldrv := rc.ResourceVersion
	rc.ResourceVersion = ""
	err = h.Update(namespace, rc)
	if err != nil {
		return err
	}

	for {
		newrc, err := h.Get(namespace, replicasetName)
		if err != nil {
			return err
		}
		if newrc.ResourceVersion == oldrv {
			time.Sleep(500 * time.Microsecond)
			continue
		}

		if *newrc.Spec.Replicas > newrc.Status.Replicas {
			for _, v := range newrc.Status.Conditions {
				if v.Type == extensionsv1beta1.ReplicaSetReplicaFailure {
					if v.Status == corev1.ConditionTrue {
						return fmt.Errorf(v.Message)
					}
				}
			}
		} else {
			return nil
		}
	}
	return nil
}

func (h *replicasetHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	rs, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(rs.Spec.Template.Labels)) {
			services = append(services, service)
		}
	}
	return services, nil
}

/*----------------- DaemonSet -----------------*/

type DaemonSetHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.DaemonSet, error)
	List(namespace string) ([]*extensionsv1beta1.DaemonSet, error)
	Create(namespace string, ds *extensionsv1beta1.DaemonSet) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *extensionsv1beta1.DaemonSet) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	Event(namespace, resourceName string) ([]corev1.Event, error)
	Revision(namespace, name string) (int64, error)
	GetRevisionsAndDescribe(namespace, name string) (map[int64]*corev1.PodTemplateSpec, error)
	Rollback(namespace, name string, revision int64) (*string, error)
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewDaemonSetHandler(group, workspace string) (DaemonSetHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &daemonsetHandler{Cluster: Cluster}, nil
}

type daemonsetHandler struct {
	*Cluster
}

func (h *daemonsetHandler) Get(namespace, name string) (*extensionsv1beta1.DaemonSet, error) {
	return h.informerController.daemonsetInformer.Lister().DaemonSets(namespace).Get(name)
}
func (h *daemonsetHandler) List(namespace string) ([]*extensionsv1beta1.DaemonSet, error) {
	return h.informerController.daemonsetInformer.Lister().DaemonSets(namespace).List(labels.Everything())
}

func (h *daemonsetHandler) Create(namespace string, daemonset *extensionsv1beta1.DaemonSet) error {
	_, err := h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Create(daemonset)
	return err
}

func (h *daemonsetHandler) Delete(namespace, daemonsetName string) error {
	return h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Delete(daemonsetName, nil)
}

func (h *daemonsetHandler) Update(namespace string, resource *extensionsv1beta1.DaemonSet) error {
	_, err := h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Update(resource)
	return err
}

func (h *daemonsetHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	d, err := h.informerController.daemonsetInformer.Lister().DaemonSets(namespace).Get(name)
	if err != nil {
		return nil, nil
	}
	rsSelector := d.Spec.Selector.MatchLabels
	selector := labels.Set(rsSelector).AsSelector()
	//opts := corev1.ListOptions{LabelSelector: selector.String()}
	//po, err := h.clientset.CoreV1().Pods(namespace).List(opts)
	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if controllerRef.UID == d.UID {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}

func (h *daemonsetHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

//参考自:k8s.io/kubernetes/pkg/kubectl/history.go
func (h *daemonsetHandler) GetControllerRevisions(namespace, name string) (*extensionsv1beta1.DaemonSet, map[int64]*appv1beta1.ControllerRevision, error) {

	d, err := h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	var allHistory []*appv1beta1.ControllerRevision
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, nil, err
	}

	historyList, err := h.clientset.AppsV1beta1().ControllerRevisions(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, nil, err
	}
	for i := range historyList.Items {
		history := historyList.Items[i]
		// Skip history that doesn't belong to the DaemonSet
		if controllerRef := metav1.GetControllerOf(&history); controllerRef == nil || controllerRef.UID != d.UID {
			continue
		}

		allHistory = append(allHistory, &history)
	}

	historyInfo := make(map[int64]*appv1beta1.ControllerRevision)
	for _, v := range allHistory {
		historyInfo[v.Revision] = v
	}

	return d, historyInfo, nil

}
func (h *daemonsetHandler) GetRevisionsAndDescribe(namespace, name string) (map[int64]*corev1.PodTemplateSpec, error) {

	d, allHistory, err := h.GetControllerRevisions(namespace, name)
	if err != nil {
		return nil, err
	}

	historySpecInfo := make(map[int64]*corev1.PodTemplateSpec)
	for _, v := range allHistory {
		//	historyInfo[v.Revision] = v
		dsOfHistory, err := applyHistory(d, v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse history %s:%v", v.Name, err)
		}
		historySpecInfo[v.Revision] = &dsOfHistory.Spec.Template
	}

	return historySpecInfo, nil

}

func (h *daemonsetHandler) Revision(namespace, name string) (int64, error) {
	d, err := h.Get(namespace, name)
	if err != nil {
		return 0, err
	}
	return d.Generation, nil

}
func applyHistory(ds *extensionsv1beta1.DaemonSet, history *appv1beta1.ControllerRevision) (*extensionsv1beta1.DaemonSet, error) {
	/*
		obj, err := k8sapi.Scheme.New(ds.GroupVersionKind())
		if err != nil {
			return nil, err
		}
		_ = obj.(*extensionsv1beta1.DaemonSet)
	*/
	clone := &extensionsv1beta1.DaemonSet{}
	cloneBytes, err := json.Marshal(clone)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(cloneBytes, history.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(patched, clone)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

//参考自:k8s.io/kubernetes/pkg/kubectl/rollback.go
func (h *daemonsetHandler) Rollback(namespace, name string, revision int64) (*string, error) {
	d, allHistory, err := h.GetControllerRevisions(namespace, name)
	if err != nil {
		return nil, err
	}
	if len(allHistory) == 0 {
		s := fmt.Sprintf("not rollout history found")
		return &s, nil
	}
	toHistory, ok := allHistory[revision]
	if !ok {
		return nil, fmt.Errorf("revision is not found")
	}

	if revision == 0 && len(allHistory) <= 1 {
		return nil, fmt.Errorf("no last revision to roll back to")
	}

	// Skip if the revision already matches current DaemonSet
	done, err := Match(d, toHistory)
	if err != nil {
		return nil, err
	}

	rollbackSkipped := "skipped rollback"

	if done {
		s := fmt.Sprintf("%s (current template already matches revision %d)", rollbackSkipped, revision)
		return &s, nil
	}
	if _, err = h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Patch(name, types.StrategicMergePatchType, toHistory.Data.Raw); err != nil {
		return nil, fmt.Errorf("failed restoring revision %d: %v", revision, err)
	}
	rollbackSuccess := "rolled back"
	return &rollbackSuccess, nil
}

func Match(ds *extensionsv1beta1.DaemonSet, history *appv1beta1.ControllerRevision) (bool, error) {
	patch, err := getPatch(ds)
	if err != nil {
		return false, err
	}
	return bytes.Equal(patch, history.Data.Raw), nil
}

func getPatch(ds *extensionsv1beta1.DaemonSet) ([]byte, error) {
	dsBytes, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	err = json.Unmarshal(dsBytes, &raw)
	if err != nil {
		return nil, err
	}
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})

	// Create a patch of the DaemonSet that replaces spec.template
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

func (h *daemonsetHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	d, err := h.informerController.daemonsetInformer.Lister().DaemonSets(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(d.Spec.Template.Labels)) {
			services = append(services, service)
		}
	}
	return services, nil
}

/* --------------- Ingress --------------*/

type IngressHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.Ingress, error)
	List(namespace string) ([]*extensionsv1beta1.Ingress, error)
	Create(namespace string, ing *extensionsv1beta1.Ingress) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *extensionsv1beta1.Ingress) error
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewIngressHandler(group, workspace string) (IngressHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &ingressHandler{Cluster: Cluster}, nil
}

type ingressHandler struct {
	*Cluster
}

func (h *ingressHandler) Get(namespace, name string) (*extensionsv1beta1.Ingress, error) {
	return h.informerController.ingressInformer.Lister().Ingresses(namespace).Get(name)
}
func (h *ingressHandler) List(namespace string) ([]*extensionsv1beta1.Ingress, error) {
	return h.informerController.ingressInformer.Lister().Ingresses(namespace).List(labels.Everything())
}

func (h *ingressHandler) Create(namespace string, ingress *extensionsv1beta1.Ingress) error {
	_, err := h.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)
	return err
}

func (h *ingressHandler) Delete(namespace string, ingressName string) error {
	return h.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, nil)
}

func (h *ingressHandler) Update(namespace string, resource *extensionsv1beta1.Ingress) error {
	_, err := h.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(resource)
	return err
}

func (h *ingressHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	ing, err := h.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	svcList := make(map[string]struct{})
	if ing.Spec.Backend != nil {
		svcList[ing.Spec.Backend.ServiceName] = struct{}{}
	}

	for _, v := range ing.Spec.Rules {
		if v.HTTP != nil {
			for _, j := range v.HTTP.Paths {
				svcList[j.Backend.ServiceName] = struct{}{}
			}
		}
	}

	svcs := make([]*corev1.Service, 0)
	for k, _ := range svcList {
		svc, err := h.informerController.serviceInformer.Lister().Services(namespace).Get(k)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, err
			} else {
				continue
			}
		}
		svcs = append(svcs, svc)
	}

	return svcs, nil
}

/* --------------- StatefulSet --------------*/

type StatefulSetHandler interface {
	Get(namespace, name string) (*appv1beta2.StatefulSet, error)
	List(namespace string) ([]*appv1beta2.StatefulSet, error)
	Create(namespace string, ss *appv1beta2.StatefulSet) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *appv1beta2.StatefulSet) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
	GetServices(namespace string, name string) ([]*corev1.Service, error)
}

func NewStatefulSetHandler(group, workspace string) (StatefulSetHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &statefulsetHandler{Cluster: Cluster}, nil
}

type statefulsetHandler struct {
	*Cluster
}

func (h *statefulsetHandler) Get(namespace, name string) (*appv1beta2.StatefulSet, error) {
	return h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).Get(name)
}
func (h *statefulsetHandler) List(namespace string) ([]*appv1beta2.StatefulSet, error) {
	return h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).List(labels.Everything())
}

func (h *statefulsetHandler) Create(namespace string, statefulset *appv1beta2.StatefulSet) error {
	_, err := h.clientset.Apps().StatefulSets(namespace).Create(statefulset)
	return err
}

func (h *statefulsetHandler) Delete(namespace, statefulsetName string) error {
	return h.clientset.Apps().StatefulSets(namespace).Delete(statefulsetName, nil)
}

func (h *statefulsetHandler) Update(namespace string, resource *appv1beta2.StatefulSet) error {
	_, err := h.clientset.Apps().StatefulSets(namespace).Update(resource)
	return err
}
func (h *statefulsetHandler) GetPods(namespace, name string) ([]*corev1.Pod, error) {
	d, err := h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).Get(name)
	if err != nil {
		return nil, nil
	}
	//	rsSelector := d.Spec.Selector.MatchLabels
	rsSelector := d.Spec.Template.Labels
	selector := labels.Set(rsSelector).AsSelector()
	//opts := corev1.ListOptions{LabelSelector: selector.String()}
	//po, err := h.clientset.CoreV1().Pods(namespace).List(opts)
	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if controllerRef.UID == d.UID {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}

func (h *statefulsetHandler) GetServices(namespace string, name string) ([]*corev1.Service, error) {
	allServices, err := h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	statefulset, err := h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	services := make([]*corev1.Service, 0)
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(statefulset.Labels)) {
			services = append(services, service)
		}
	}
	return services, nil
}

/* --------------- CronJob --------------*/

type CronJobHandler interface {
	Get(namespace, name string) (*batchv2alpha1.CronJob, error)
	List(namespace string) ([]*batchv2alpha1.CronJob, error)
	Create(namespace string, cj *batchv2alpha1.CronJob) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *batchv2alpha1.CronJob) error
	GetJobs(namespace, name string) ([]*batchv1.Job, error)
	Event(namespace, resourceName string) ([]corev1.Event, error)
}

func NewCronJobHandler(group, workspace string) (CronJobHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &cronjobHandler{Cluster: Cluster, Group: group}, nil
}

type cronjobHandler struct {
	*Cluster
	Group string
}

func (h *cronjobHandler) Get(namespace, name string) (*batchv2alpha1.CronJob, error) {
	return h.informerController.cronjobInformer.Lister().CronJobs(namespace).Get(name)
}

func (h *cronjobHandler) List(namespace string) ([]*batchv2alpha1.CronJob, error) {
	return h.informerController.cronjobInformer.Lister().CronJobs(namespace).List(labels.Everything())
}
func (h *cronjobHandler) Create(namespace string, cronjob *batchv2alpha1.CronJob) error {
	_, err := h.clientset.BatchV2alpha1().CronJobs(namespace).Create(cronjob)
	return err
}

func (h *cronjobHandler) Delete(namespace, cronjobName string) error {
	jobs, err := h.GetJobs(namespace, cronjobName)
	if err != nil {
		return err
	}

	err = h.clientset.BatchV2alpha1().CronJobs(namespace).Delete(cronjobName, nil)
	if err != nil {
		return err
	}
	if len(jobs) != 0 {
		jh, err := NewJobHandler(h.Group, namespace)
		if err != nil {
			return err
		}
		for _, v := range jobs {
			err := jh.Delete(namespace, v.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *cronjobHandler) Update(namespace string, resource *batchv2alpha1.CronJob) error {
	_, err := h.clientset.BatchV2alpha1().CronJobs(namespace).Update(resource)
	return err
}

func (h *cronjobHandler) GetJobs(namespace, cronjobName string) ([]*batchv1.Job, error) {
	cj, err := h.Get(namespace, cronjobName)
	if err != nil {
		return nil, err
	}

	jobs := make([]*batchv1.Job, 0)
	//不存在标签的cronjob,无法获取其相关的Pod
	if len(cj.Spec.JobTemplate.Labels) != 0 {
		//	log.DebugPrint(cj.Spec.JobTemplate.Labels)
		//	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
		//selector := fields.SelectorFromSet(cj.Spec.JobTemplate.Labels)
		selector := labels.SelectorFromSet(cj.Spec.JobTemplate.Labels)
		alljobs, err := h.informerController.jobInformer.Lister().Jobs(namespace).List(selector)
		if err != nil {
			return nil, err
		}

		for k := range alljobs {
			controllerRef := metav1.GetControllerOf(alljobs[k])
			if controllerRef == nil {
				continue
			}

			if controllerRef.UID == cj.UID {
				jobs = append(jobs, alljobs[k])
			}
		}
	}
	sort.Sort(SortableJobs(jobs))
	return jobs, nil
}

func (h *cronjobHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

/* ------------------------- Job ---------------------*/
type JobHandler interface {
	Get(namespace, name string) (*batchv1.Job, error)
	List(namespace string) ([]*batchv1.Job, error)

	Create(namespace string, job *batchv1.Job) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *batchv1.Job) error
	GetPods(Namespace, name string) ([]*corev1.Pod, error)
	Event(namespace, resourceName string) ([]corev1.Event, error)
	GetCreator(namespace, name string) (*corev1.SerializedReference, error)
}

func NewJobHandler(group, workspace string) (JobHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &jobHandler{Cluster: Cluster}, nil
}

type jobHandler struct {
	*Cluster
}

func (h *jobHandler) Get(namespace, name string) (*batchv1.Job, error) {
	return h.informerController.jobInformer.Lister().Jobs(namespace).Get(name)
}
func (h *jobHandler) List(namespace string) ([]*batchv1.Job, error) {
	return h.informerController.jobInformer.Lister().Jobs(namespace).List(labels.Everything())
}

func (h *jobHandler) Create(namespace string, job *batchv1.Job) error {
	_, err := h.clientset.BatchV1().Jobs(namespace).Create(job)
	return err
}

func (h *jobHandler) Delete(namespace, jobName string) error {
	job, err := h.Get(namespace, jobName)
	if err != nil {
		return err
	}
	err = h.clientset.BatchV1().Jobs(namespace).Delete(jobName, nil)
	if err != nil {
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		err2 := fmt.Errorf("clean job pod %v fail for %v,please clean them by using kubectl command", jobName, err)
		return err2
	}
	opt := metav1.ListOptions{LabelSelector: selector.String()}
	err = h.clientset.CoreV1().Pods(namespace).DeleteCollection(nil, opt)
	if err != nil {
		err2 := fmt.Errorf("clean job pod %v fail for %v,please clean them by using kubectl command", jobName, err)
		return err2
	}

	return nil
}

func (h *jobHandler) Update(namespace string, resource *batchv1.Job) error {
	_, err := h.clientset.BatchV1().Jobs(namespace).Update(resource)
	return err
}

func (h *jobHandler) GetPods(namespace, jobName string) ([]*corev1.Pod, error) {
	job, err := h.informerController.jobInformer.Lister().Jobs(namespace).Get(jobName)
	if err != nil {
		return nil, err
	}

	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, err
	}

	allpos, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	pos := make([]*corev1.Pod, 0)
	for k := range allpos {
		controllerRef := metav1.GetControllerOf(allpos[k])
		if controllerRef == nil {
			continue
		}

		if controllerRef.UID == job.UID {
			pos = append(pos, allpos[k])
		}
	}

	return pos, nil
}
func (h *jobHandler) Event(namespace, resourceName string) ([]corev1.Event, error) {
	//	pod, err := h.clientset.Pods(namespace).Get(podName, metav1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&resourceName, &namespace, nil, nil)
	options := metav1.ListOptions{FieldSelector: selector.String()}
	events, err2 := h.clientset.CoreV1().Events(namespace).List(options)
	if err2 != nil {
		return nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	return events.Items, nil
}

func (h *jobHandler) GetCreator(namespace string, name string) (*corev1.SerializedReference, error) {
	pod, err := h.informerController.jobInformer.Lister().Jobs(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	sr, err := getResourceCreator(pod)
	if err != nil {
		return nil, err
	}

	if sr == nil {
		return nil, nil
	}
	v1sr := kubernetesapiSerrializedReferenceToClientGo(*sr)

	return &v1sr, nil
}

type HorizontalPodAutoscalerHandler interface {
	Get(namespace string, name string) (*autoscalingv1.HorizontalPodAutoscaler, error)
	Delete(namespace string, name string) error
	Create(namespace string, hpa *autoscalingv1.HorizontalPodAutoscaler) error
	Update(namespace string, hpa *autoscalingv1.HorizontalPodAutoscaler) error
	List(namespace string) ([]*autoscalingv1.HorizontalPodAutoscaler, error)
}

func NewHorizontalPodAutoscalerHandler(group, workspace string) (HorizontalPodAutoscalerHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &hpaHandler{Cluster: Cluster}, nil
}

type hpaHandler struct {
	*Cluster
}

func (h *hpaHandler) Get(namespace, name string) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	return h.informerController.hpaInformer.Lister().HorizontalPodAutoscalers(namespace).Get(name)
}

func (h *hpaHandler) Create(namespace string, hpa *autoscalingv1.HorizontalPodAutoscaler) error {
	_, err := h.clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Create(hpa)
	return err
}

func (h *hpaHandler) Update(namespace string, hpa *autoscalingv1.HorizontalPodAutoscaler) error {
	_, err := h.clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Update(hpa)
	return err
}

func (h *hpaHandler) Delete(namespace, hpaName string) error {
	return h.clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Delete(hpaName, nil)
}

func (h *hpaHandler) List(namespace string) ([]*autoscalingv1.HorizontalPodAutoscaler, error) {
	return h.informerController.hpaInformer.Lister().HorizontalPodAutoscalers(namespace).List(labels.Everything())
}

/*  helpers */

func watchRollbackEvent(w watch.Interface) string {
	for {
		select {
		case event, ok := <-w.ResultChan():
			if !ok {
				return ""
			}
			obj, ok := event.Object.(*corev1.Event)
			if !ok {
				w.Stop()
				return ""
			}
			isRollback, result := isRollbackEvent(obj)
			if isRollback {
				w.Stop()
				return result
			}
			//		case <-signals:
			//			w.Stop()
		}
	}
}

func isRollbackEvent(e *corev1.Event) (bool, string) {
	rollbackEventReasons := []string{deploymentutil.RollbackRevisionNotFound, deploymentutil.RollbackTemplateUnchanged, deploymentutil.RollbackDone}
	for _, reason := range rollbackEventReasons {
		if e.Reason == reason {
			if reason == deploymentutil.RollbackDone {
				return true, "rolled back"
			}
			return true, fmt.Sprintf("skipped rollback (%s:%s)", e.Reason, e.Message)
		}
	}
	return false, ""
}

func getResourceCreator(obj interface{}) (*kubernetesapi.SerializedReference, error) {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	var creatorRef string
	var found bool
	creatorRef, found = accessor.GetAnnotations()[kubernetesapi.CreatedByAnnotation]
	if !found {
		return nil, nil
	}

	decoder := kubernetesapi.Codecs.UniversalDecoder()
	sr := &kubernetesapi.SerializedReference{}
	if err := runtime.DecodeInto(decoder, []byte(creatorRef), sr); err != nil {
		return nil, err
	}

	return sr, nil
}

func kubernetesapiSerrializedReferenceToClientGo(sr kubernetesapi.SerializedReference) corev1.SerializedReference {
	var apisr corev1.SerializedReference
	apisr.APIVersion = sr.APIVersion
	apisr.Kind = sr.Kind
	apisr.Reference.APIVersion = sr.Reference.APIVersion
	apisr.Reference.FieldPath = sr.Reference.FieldPath
	apisr.Reference.Kind = sr.Reference.Kind
	apisr.Reference.Name = sr.Reference.Name
	apisr.Reference.Namespace = sr.Reference.Namespace
	apisr.Reference.ResourceVersion = sr.Reference.ResourceVersion
	apisr.Reference.UID = sr.Reference.UID
	return apisr

}

func runtimeObjectListToObjectReference(obj interface{}) []corev1.ObjectReference {

	ors := make([]corev1.ObjectReference, 0)
	switch res := obj.(type) {
	case []*appv1beta1.StatefulSet:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "StatefulSet"
			or.APIVersion = "apps/v1beta1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*extensionsv1beta1.ReplicaSet:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "ReplicaSet"
			or.APIVersion = "extensions/v1beta1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*corev1.ReplicationController:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "ReplicationController"
			or.APIVersion = "v1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*corev1.Pod:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "Pod"
			or.APIVersion = "v1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			or.UID = v.UID
			ors = append(ors, or)
		}
	case []*extensionsv1beta1.Deployment:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "Deployment"
			or.APIVersion = "extensions/v1beta1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*extensionsv1beta1.DaemonSet:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "DaemonSet"
			or.APIVersion = "extensions/v1beta1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*batchv2alpha1.CronJob:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "CronJob"
			or.APIVersion = "batch/v2alpha1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	case []*batchv1.Job:
		for _, v := range res {
			var or corev1.ObjectReference
			or.Kind = "Job"
			or.APIVersion = "batch/v1"
			or.Name = v.Name
			or.ResourceVersion = v.ResourceVersion
			or.Namespace = v.Namespace
			ors = append(ors, or)
		}
	default:
		log.ErrorPrint("unsupport type")
	}
	return ors
}

func IsPodSpecReferenceConfigMap(name string, spec corev1.PodSpec) bool {
	for _, c := range spec.Containers {
		for _, j := range c.EnvFrom {
			if j.ConfigMapRef != nil {
				if j.ConfigMapRef.Name == name {
					goto found
				}
			}
		}
		for _, j := range c.Env {

			if j.ValueFrom != nil {
				if j.ValueFrom.ConfigMapKeyRef != nil {
					if j.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name == name {
						goto found
					}
				}

			}
		}
	}
	return false
	//已找到,无须进行其他的遍历
found:
	return true
}

func IsPodSpecReferenceSecret(name string, spec corev1.PodSpec) bool {
	for _, c := range spec.Containers {
		for _, j := range c.Env {

			if j.ValueFrom != nil {
				if j.ValueFrom.ConfigMapKeyRef != nil {
					if j.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name == name {
						goto found
					}
				}

			}
		}
	}
	return false
	//已找到,无须进行其他的遍历
found:
	return true
}

func IsPodSpecReferenceServiceAccount(name string, spec corev1.PodSpec) bool {
	if spec.ServiceAccountName == name {
		return true
	}
	return false
}

type podspecReferencCheckFn func(name string, spec corev1.PodSpec) bool

func getGeneralResourceReference(informerController *ResourceController, namespace string, name string, fn podspecReferencCheckFn) ([]corev1.ObjectReference, error) {
	ors := make([]corev1.ObjectReference, 0)

	//statefulset
	allsfs, err := informerController.statefulsetInformer.Lister().StatefulSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	//daemonset,需要使用的是Template的
	alldts, err := informerController.daemonsetInformer.Lister().DaemonSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	//deployment
	allds, err := informerController.deploymentInformer.Lister().Deployments(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	//replicaset
	allrss, err := informerController.replicasetInformer.Lister().ReplicaSets(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	//replicationcontroller
	allrcs, err := informerController.replicationcontrollerInformer.Lister().ReplicationControllers(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	allcronjobs, err := informerController.cronjobInformer.Lister().CronJobs(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	alljobs, err := informerController.jobInformer.Lister().Jobs(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	//pod
	allps, err := informerController.podInformer.Lister().Pods(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	sfs := make([]*appv1beta2.StatefulSet, 0)
	for _, v := range allsfs {
		found := fn(name, v.Spec.Template.Spec)
		if found {
			sfs = append(sfs, v)
		}
	}
	dts := make([]*extensionsv1beta1.DaemonSet, 0)
	for _, v := range alldts {
		found := fn(name, v.Spec.Template.Spec)
		if found {
			dts = append(dts, v)
		}
	}

	ds := make([]*extensionsv1beta1.Deployment, 0)
	for _, v := range allds {
		found := fn(name, v.Spec.Template.Spec)
		if found {
			ds = append(ds, v)
		}
	}

	rss := make([]*extensionsv1beta1.ReplicaSet, 0)
	for _, v := range allrss {
		found := fn(name, v.Spec.Template.Spec)
		if found {
			rss = append(rss, v)
		}
	}

	rcs := make([]*corev1.ReplicationController, 0)
	for _, v := range allrcs {
		if v.Spec.Template != nil {
			found := fn(name, v.Spec.Template.Spec)
			if found {
				rcs = append(rcs, v)
			}
		}
	}

	cronjobs := make([]*batchv2alpha1.CronJob, 0)
	for _, v := range allcronjobs {
		found := fn(name, v.Spec.JobTemplate.Spec.Template.Spec)
		if found {
			cronjobs = append(cronjobs, v)
		}
	}

	jobs := make([]*batchv1.Job, 0)
	for _, v := range alljobs {
		found := fn(name, v.Spec.Template.Spec)
		if found {
			jobs = append(jobs, v)
		}
	}

	ps := make([]*corev1.Pod, 0)
	for _, v := range allps {
		found := fn(name, v.Spec)
		if found {
			ps = append(ps, v)
		}
	}

	ors = append(ors, runtimeObjectListToObjectReference(sfs)...)
	ors = append(ors, runtimeObjectListToObjectReference(dts)...)
	ors = append(ors, runtimeObjectListToObjectReference(ds)...)
	ors = append(ors, runtimeObjectListToObjectReference(rss)...)
	ors = append(ors, runtimeObjectListToObjectReference(rcs)...)
	ors = append(ors, runtimeObjectListToObjectReference(cronjobs)...)
	ors = append(ors, runtimeObjectListToObjectReference(jobs)...)
	ors = append(ors, runtimeObjectListToObjectReference(ps)...)
	return ors, nil

}
