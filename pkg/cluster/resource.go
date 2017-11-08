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
	"k8s.io/kubernetes/pkg/controller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	externalclientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	core "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/core/v1"
	extensions "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/meta"
	watch "k8s.io/apimachinery/pkg/watch"
	kubernetesapi "k8s.io/kubernetes/pkg/api"
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
	/*
			switch sr.Reference.Kind {
			case "ReplicationController":
				return h.informerController.replicasetInformer.Lister().ReplicaSets(sr.Reference.Namespace).Get(sr.Reference.Name)
			case "DaemonSet":
				return h.informerController.daemonsetInformer.Lister().DaemonSets(sr.Reference.Namespace).Get(sr.Reference.Name)
			case "Job":
				return h.informerController.jobInformer.Lister().Jobs(sr.Reference.Namespace).Get(sr.Reference.Name)
			case "ReplicaSet":
				return h.informerController.replicasetInformer.Lister().ReplicaSets(sr.Reference.Namespace).Get(sr.Reference.Name)
			case "StatefulSet":
				return h.informerController.statefulsetInformer.Lister().StatefulSets(sr.Reference.Namespace).Get(sr.Reference.Name)
			}
		return nil, fmt.Errorf("Unknown controller kind %q", sr.Reference.Kind)
	*/

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
	pods, err := h.informerController.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (h *serviceHandler) List(namespace string) ([]*corev1.Service, error) {
	return h.informerController.serviceInformer.Lister().Services(namespace).List(labels.Everything())
}

/* ------------------------ Configmap ----------------------------*/

type ConfigMapHandler interface {
	Get(namespace string, name string) (*corev1.ConfigMap, error)
	Create(namespace string, cm *corev1.ConfigMap) error
	Delete(namespace string, name string) error
	Update(namespace string, service *corev1.ConfigMap) error
	List(namespace string) ([]*corev1.ConfigMap, error)
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
	pos, err := h.informerController.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
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

/* ------------------------ Secret ----------------------------*/

type SecretHandler interface {
	Get(namespace string, name string) (*corev1.Secret, error)
	Create(namespace string, secret *corev1.Secret) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *corev1.Secret) error
	List(namespace string) ([]*corev1.Secret, error)
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
	d, err := h.Get(namespace, deploymentName)
	if err != nil {
		return err
	}
	allOldRSs, newRs, err := h.GetAllReplicaSets(namespace, deploymentName)
	allRSs := allOldRSs
	if newRs != nil {
		allRSs = append(allRSs, newRs)
	}

	err = h.clientset.ExtensionsV1beta1().Deployments(namespace).Delete(deploymentName, nil)
	if err != nil {
		return err
	}

	for _, v := range allRSs {
		err := h.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Delete(v.Name, nil)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.ErrorPrint(fmt.Sprintf("try to delete rs %v fail for %v", v.Name, err))
			}
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		log.ErrorPrint(fmt.Sprintf("try to delete pods fail for %v", err))
	}

	opt := metav1.ListOptions{LabelSelector: selector.String()}
	err = h.clientset.CoreV1().Pods(namespace).DeleteCollection(nil, opt)
	if err != nil {
		log.ErrorPrint(fmt.Sprintf("try to delete pods fail for %v", err))
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
	pos, err := h.informerController.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
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

	event, err := h.clientset.Events(namespace).List(metav1.ListOptions{})
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

func (h *deploymentHandler) GetAllReplicaSets(namespace string, name string) ([]*extensionsv1beta1.ReplicaSet, *extensionsv1beta1.ReplicaSet, error) {
	//func (h *deploymentHandler) GetAllReplicaSets(namespace string, name string) ([]*k8sextensions.ReplicaSet, *k8sextensions.ReplicaSet, error) {

	internalExtensionClientset := extensions.New(h.clientset.ExtensionsV1beta1().RESTClient())
	internalCoreClientset := core.New(h.clientset.CoreV1().RESTClient())
	versionedClient := &externalclientset.Clientset{
		CoreV1Client:            internalCoreClientset,
		ExtensionsV1beta1Client: internalExtensionClientset,
	}

	deployment, err := versionedClient.Extensions().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve deployment %s: %v", name, err)
	}
	_, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, versionedClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve replica sets from deployment %s: %v", name, err)
	}

	clientAllOldRSs := make([]*extensionsv1beta1.ReplicaSet, 0)
	for _, v := range allOldRSs {
		rs, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).Get(v.Name)
		if err != nil {
			return nil, nil, err
		}
		clientAllOldRSs = append(clientAllOldRSs, rs)
	}

	clientNewRSs, err := h.informerController.replicasetInformer.Lister().ReplicaSets(namespace).Get(newRS.Name)
	if err != nil {
		return nil, nil, err
	}

	return clientAllOldRSs, clientNewRSs, nil
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
	pos, err := h.informerController.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
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
	pos, err := h.informerController.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
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
		if controllerRef := controller.GetControllerOf(&history); controllerRef == nil || controllerRef.UID != d.UID {
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

/* --------------- StatefulSet --------------*/

type StatefulSetHandler interface {
	Get(namespace, name string) (*appv1beta1.StatefulSet, error)
	List(namespace string) ([]*appv1beta1.StatefulSet, error)
	Create(namespace string, ss *appv1beta1.StatefulSet) error
	Delete(namespace string, name string) error
	Update(namespace string, resource *appv1beta1.StatefulSet) error
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

func (h *statefulsetHandler) Get(namespace, name string) (*appv1beta1.StatefulSet, error) {
	return h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).Get(name)
}
func (h *statefulsetHandler) List(namespace string) ([]*appv1beta1.StatefulSet, error) {
	return h.informerController.statefulsetInformer.Lister().StatefulSets(namespace).List(labels.Everything())
}

func (h *statefulsetHandler) Create(namespace string, statefulset *appv1beta1.StatefulSet) error {
	_, err := h.clientset.Apps().StatefulSets(namespace).Create(statefulset)
	return err
}

func (h *statefulsetHandler) Delete(namespace, statefulsetName string) error {
	return h.clientset.Apps().StatefulSets(namespace).Delete(statefulsetName, nil)
}

func (h *statefulsetHandler) Update(namespace string, resource *appv1beta1.StatefulSet) error {
	_, err := h.clientset.AppsV1beta1().StatefulSets(namespace).Update(resource)
	return err
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
	for _, v := range jobs {
		log.DebugPrint(v.Name)
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
		var err error
		jobs, err = h.informerController.jobInformer.Lister().Jobs(namespace).List(selector)
		if err != nil {
			return nil, err
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
	err = h.clientset.Pods(namespace).DeleteCollection(nil, opt)
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

	pods, err := h.informerController.podInformer.Lister().Pods(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	return pods, nil
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
