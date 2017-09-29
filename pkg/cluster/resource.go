package cluster

import (
	"fmt"
	"ufleet-deploy/pkg/log"

	corev1 "k8s.io/client-go/pkg/api/v1"
	appv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

//Get从Informer中获取
//Delete/Create则调用k8s相应的接口

/* ----------------- Pod ----------------------*/
type PodHandler interface {
	Get(namespace string, name string) (*corev1.Pod, error)
	Delete(namespace string, name string) error
	Create(namespace string, pod *corev1.Pod) error
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

func (h *podHandler) Get(namespace, name string) (*corev1.Pod, error) {
	return h.informerController.podInformer.Lister().Pods(namespace).Get(name)
}

func (h *podHandler) Create(namespace string, pod *corev1.Pod) error {
	_, err := h.clientset.CoreV1().Pods(namespace).Create(pod)
	return err
}

func (h *podHandler) Delete(namespace, podName string) error {
	return h.clientset.CoreV1().Pods(namespace).Delete(podName, nil)
}

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
}

func NewCronJobHandler(group, workspace string) (CronJobHandler, error) {
	Cluster, err := Controller.GetCluster(group, workspace)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return &cronjobHandler{Cluster: Cluster}, nil
}

type cronjobHandler struct {
	*Cluster
}

func (h *cronjobHandler) Get(namespace, name string) (*batchv2alpha1.CronJob, error) {
	return h.informerController.cronjobInformer.Lister().CronJobs(namespace).Get(name)
}

func (h *cronjobHandler) Create(namespace string, cronjob *batchv2alpha1.CronJob) error {
	_, err := h.clientset.BatchV2alpha1().CronJobs(namespace).Create(cronjob)
	return err
}

func (h *cronjobHandler) Delete(namespace, cronjobName string) error {
	return h.clientset.BatchV2alpha1().CronJobs(namespace).Delete(cronjobName, nil)
}

/* ------------------------- Job ---------------------*/
type JobHandler interface {
	Get(namespace, name string) (*batchv1.Job, error)
	Create(namespace string, job *batchv1.Job) error
	Delete(namespace string, name string) error
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
	return h.clientset.BatchV1().Jobs(namespace).Delete(jobName, nil)
}
