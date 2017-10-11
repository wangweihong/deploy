package cluster

import (
	"fmt"
	"sort"
	"time"
	"ufleet-deploy/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
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

func (h *podHandler) GetFromClientSet(namespace, name string) (*corev1.Pod, error) {
	//	return h.informerController.podInformer.Lister().Pods(namespace).Get(name)
	return h.clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func (h *podHandler) Get(namespace, name string, opt GetOptions) (*corev1.Pod, error) {
	if opt.Direct {
		pod, err := h.clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		//	log.DebugPrint(pod.TypeMeta)
		//	log.DebugPrint(pod.ObjectMeta)
		//	log.DebugPrint(pod.APIVersion)
		return pod, nil
	}
	return h.informerController.podInformer.Lister().Pods(namespace).Get(name)
}

func (h *podHandler) Create(namespace string, pod *corev1.Pod) error {
	_, err := h.clientset.CoreV1().Pods(namespace).Create(pod)
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

/*
func (h *podHandler) Event(namespace, podName string) (string, []corev1.Event, error) {
	pod, err := h.clientset.Pods(namespace).Get(podName, meta_v1.GetOptions{})
	selector := h.clientset.CoreV1().Events(namespace).GetFieldSelector(&podName,&namespace,nil,nil)
	options := corev1.ListOptions{FieldSelector: selector.String())
	if err2!=nil {
		return "", nil, err2
	}

	//获取不到Pod,但有Pod事件
	sort.Sort(SortableEvents(events.Items))
	if err != nil {
		return "", events.Items,nil
	}
}
*/

/* ----------------- Service ----------------------*/

type ServiceHandler interface {
	Get(namespace string, name string) (*corev1.Service, error)
	Delete(namespace string, name string) error
	Create(namespace string, service *corev1.Service) error
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

func (h *serviceHandler) Delete(namespace, serviceName string) error {
	return h.clientset.CoreV1().Services(namespace).Delete(serviceName, nil)
}

/* ------------------------ Configmap ----------------------------*/

type ConfigMapHandler interface {
	Get(namespace string, name string) (*corev1.ConfigMap, error)
	Create(namespace string, cm *corev1.ConfigMap) error
	Delete(namespace string, name string) error
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

func (h *configmapHandler) Delete(namespace, configmapName string) error {
	return h.clientset.CoreV1().ConfigMaps(namespace).Delete(configmapName, nil)
}

/* ------------------------ ReplicationController---------------------------*/

type ReplicationControllerHandler interface {
	Get(namespace string, name string) (*corev1.ReplicationController, error)
	Create(namespace string, cm *corev1.ReplicationController) error
	Delete(namespace string, name string) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
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

/* ------------------------ ServiceAccount ----------------------------*/

type ServiceAccountHandler interface {
	Get(namespace string, name string) (*corev1.ServiceAccount, error)
	Create(namespace string, sa *corev1.ServiceAccount) error
	Delete(namespace string, name string) error
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

func (h *serviceaccountHandler) Delete(namespace, serviceaccountName string) error {
	return h.clientset.CoreV1().ServiceAccounts(namespace).Delete(serviceaccountName, nil)
}

/* ------------------------ Secret ----------------------------*/

type SecretHandler interface {
	Get(namespace string, name string) (*corev1.Secret, error)
	Create(namespace string, secret *corev1.Secret) error
	Delete(namespace string, name string) error
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

/* ------------------------ Endpoint ----------------------------*/

type EndpointHandler interface {
	Get(namespace string, name string) (*corev1.Endpoints, error)
	Create(namespace string, ep *corev1.Endpoints) error
	Delete(namespace string, name string) error
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

/* ------------------------ Deployment ----------------------------*/

type DeploymentHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.Deployment, error)
	Create(namespace string, d *extensionsv1beta1.Deployment) error
	Delete(namespace string, name string) error
	Scale(namespace, name string, num int32) error
	GetPods(namespace, name string) ([]*corev1.Pod, error)
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

func (h *deploymentHandler) Create(namespace string, deployment *extensionsv1beta1.Deployment) error {
	_, err := h.clientset.ExtensionsV1beta1().Deployments(namespace).Create(deployment)
	return err
}

func (h *deploymentHandler) Delete(namespace, deploymentName string) error {
	return h.clientset.ExtensionsV1beta1().Deployments(namespace).Delete(deploymentName, nil)
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

/*----------------- DaemonSet -----------------*/

type DaemonSetHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.DaemonSet, error)
	Create(namespace string, ds *extensionsv1beta1.DaemonSet) error
	Delete(namespace string, name string) error
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

func (h *daemonsetHandler) Create(namespace string, daemonset *extensionsv1beta1.DaemonSet) error {
	_, err := h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Create(daemonset)
	return err
}

func (h *daemonsetHandler) Delete(namespace, daemonsetName string) error {
	return h.clientset.ExtensionsV1beta1().DaemonSets(namespace).Delete(daemonsetName, nil)
}

/* --------------- Ingress --------------*/

type IngressHandler interface {
	Get(namespace string, name string) (*extensionsv1beta1.Ingress, error)
	Create(namespace string, ing *extensionsv1beta1.Ingress) error
	Delete(namespace string, name string) error
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

func (h *ingressHandler) Create(namespace string, ingress *extensionsv1beta1.Ingress) error {
	_, err := h.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)
	return err
}

func (h *ingressHandler) Delete(namespace string, ingressName string) error {
	return h.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, nil)
}

/* --------------- StatefulSet --------------*/

type StatefulSetHandler interface {
	Get(namespace, name string) (*appv1beta1.StatefulSet, error)
	Create(namespace string, ss *appv1beta1.StatefulSet) error
	Delete(namespace string, name string) error
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

func (h *statefulsetHandler) Create(namespace string, statefulset *appv1beta1.StatefulSet) error {
	_, err := h.clientset.Apps().StatefulSets(namespace).Create(statefulset)
	return err
}

func (h *statefulsetHandler) Delete(namespace, statefulsetName string) error {
	return h.clientset.Apps().StatefulSets(namespace).Delete(statefulsetName, nil)
}

/* --------------- CronJob --------------*/

type CronJobHandler interface {
	Get(namespace, name string) (*batchv2alpha1.CronJob, error)
	Create(namespace string, cj *batchv2alpha1.CronJob) error
	Delete(namespace string, name string) error
	GetJobs(namespace, name string) ([]*batchv1.Job, error)
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

/* ------------------------- Job ---------------------*/
type JobHandler interface {
	Get(namespace, name string) (*batchv1.Job, error)
	Create(namespace string, job *batchv1.Job) error
	Delete(namespace string, name string) error
	GetPods(Namespace, name string) ([]*corev1.Pod, error)
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

/*  helpers */

type SortableEvents []corev1.Event

func (list SortableEvents) Len() int {
	return len(list)
}

func (list SortableEvents) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableEvents) Less(i, j int) bool {
	return list[i].LastTimestamp.Time.Before(list[j].LastTimestamp.Time)
}

/*job*/

type SortableJobs []*batchv1.Job

func (list SortableJobs) Len() int {
	return len(list)
}

func (list SortableJobs) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

//按距离当前的创建时间进行排序
func (list SortableJobs) Less(i, j int) bool {
	return list[i].CreationTimestamp.Time.After(list[j].CreationTimestamp.Time)
}
