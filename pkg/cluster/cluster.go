package cluster

import (
	"fmt"
	"sync"
	"time"
	"ufleet-deploy/pkg/log"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

var (
	ErrClusterNotFound = fmt.Errorf("cluster not found")
	ErrClusterExists   = fmt.Errorf("cluster already exists")

	Controller ClusterController = &clusterController{
		cis:      make(map[string]*Cluster),
		clusters: make(map[string]Cluster),
		locker:   sync.Mutex{},
	}
)

type ClusterController interface {
	CreateCluster(group, workspace string) error
	//只有在集群全部相关的workspace都被移除时,cluster才会真正被移除
	DeleteCluster(group, workspace string) error
	GetCluster(group, workspace string) (*Cluster, error)
}

type Workspace struct {
	Name  string
	Group string
}

//李强删除集群的时候,会调用接口删除集群上的应用
//在调用的时候再关闭informoers
func (c *Cluster) CloseInformers() {
	c.informerStopChan <- struct{}{}
}

func (c *Cluster) StartInformers() error {

	rclient, err := kubernetes.NewForConfig(c.Config)
	if err != nil {
		return log.DebugPrint(err)
	}
	c.clientset = rclient

	//每隔60分钟,触发一次Update事件
	sharedInformerFactory := informers.NewSharedInformerFactory(rclient, 60*time.Second)
	controller := NewResourceController(sharedInformerFactory, c.Workspaces)
	err = controller.Run(c.informerStopChan)
	if err != nil {
		return log.DebugPrint(err)
	}
	c.informerController = controller

	return nil
}

type clusterController struct {
	//根据组_工作区记录集群
	cis map[string]*Cluster //key:"group_workspace", Cluster
	//根据集群名记录集群
	clusters map[string]Cluster //key: clusterName
	locker   sync.Mutex
}

func (c *clusterController) CreateCluster(group, workspace string) error {

	c.locker.Lock()
	defer c.locker.Unlock()

	gwkey := group + "_" + workspace
	_, ok := c.cis[gwkey]
	if ok {
		return ErrClusterExists
	}

	token := "1234567890987654321"
	cConfig, err := GetK8sClientConfig(group, workspace, token)
	if err != nil {
		return log.DebugPrint(err)
	}

	rconfig, err := ClusterConfigToK8sClientConfig(*cConfig)
	if err != nil {
		return log.DebugPrint(err)
	}

	cluster, ok := c.clusters[cConfig.ClusterName]
	if ok {
		cluster.Reference += 1
		cluster.Workspaces[workspace] = Workspace{Name: workspace, Group: group}
		c.cis[gwkey] = &cluster
		c.clusters[cConfig.ClusterName] = cluster
		return nil
	} else {
		log.DebugPrint("start to create cluster :%v", cConfig.ClusterName)
		var cluster Cluster
		cluster.Name = cConfig.ClusterName
		cluster.Workspaces = make(map[string]Workspace)
		cluster.informerStopChan = make(chan struct{})
		cluster.Reference = 1
		cluster.Config = rconfig
		cluster.Workspaces[workspace] = Workspace{Name: workspace, Group: group}

		err := cluster.StartInformers()
		if err != nil {
			return log.DebugPrint(err)
		}
		//

		c.clusters[cConfig.ClusterName] = cluster
		c.cis[gwkey] = &cluster
	}

	return nil
}

func (c *clusterController) GetCluster(group, workspace string) (*Cluster, error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	key := group + "_" + workspace
	//	ci, ok := groupWorkspaceToCluster[key]
	ci, ok := c.cis[key]
	if !ok {
		return nil, ErrClusterNotFound
	}
	return ci, nil

}
func (c *clusterController) DeleteCluster(group, workspace string) error {

	c.locker.Lock()
	defer c.locker.Unlock()

	key := group + "_" + workspace
	pcluster, ok := c.cis[key]
	if !ok {
		return ErrClusterNotFound
	}
	delete(c.cis, key)

	cluster := c.clusters[pcluster.Name]
	cluster.Reference -= 1

	//已经没有工作区了
	//停止cluster,即使后续会往集群中添加新的工作区
	if cluster.Reference == 0 {
		log.DebugPrint("start to delete cluster :%v", cluster.Name)
		cluster.CloseInformers()
		delete(c.clusters, pcluster.Name)
		//仍有工作区,仅仅清除组/工作区信息
	} else {
		c.clusters[pcluster.Name] = cluster
	}

	return nil
}

type Cluster struct {
	Name             string `json:"name"`
	Workspaces       map[string]Workspace
	informerStopChan chan struct{} //集群上各资源的informer stop chan
	Locker           sync.Mutex
	Config           *rest.Config //TODO:如何在证书修改时更新证书
	Reference        int

	informerController *ResourceController
	clientset          *kubernetes.Clientset
}
