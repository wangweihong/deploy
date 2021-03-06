package cluster

import (
	"fmt"
	"strings"
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

	globalClusterController = clusterController{
		cis:      make(map[string]*Cluster),
		clusters: make(map[string]*Cluster),
		locker:   sync.Mutex{},
	}
	Controller ClusterController = &globalClusterController
)

type ClusterController interface {
	CreateOrUpdateCluster(group, workspace string, startinformer bool) (*Cluster, bool, error)
	//只有在集群全部相关的workspace都被移除时,cluster才会真正被移除
	DeleteCluster(group, workspace string) error
	GetCluster(group, workspace string) (*Cluster, error)
	GetClusterIP(group, workspace string) (*string, error)
	GetClusterVIP(group, workspace string) (*string, error)
}

type Workspace struct {
	Name  string
	Group string
}

//李强删除集群的时候,会调用接口删除集群上的应用
//在调用的时候再关闭informoers
func (c *Cluster) ill() error {
	rclient, err := kubernetes.NewForConfig(c.Config)
	if err != nil {
		return err
	}
	_, err = rclient.ServerVersion()
	if err != nil {
		return err

	}
	return nil
}

func (c *Cluster) CloseInformers() {
	if c.informerStart {
		//	c.informerStopChan <- struct{}{}
		close(c.informerStopChan)
		c.informerStart = false
	}
}

func (c *Cluster) StartInformers() error {
	if c.informerStart {
		return nil
	}
	rclient, err := kubernetes.NewForConfig(c.Config)
	if err != nil {
		return log.DebugPrint(err)
	}
	c.clientset = rclient

	//每隔60分钟,触发一次Update事件
	sharedInformerFactory := informers.NewSharedInformerFactory(rclient, 60*time.Hour)
	controller := NewResourceController(sharedInformerFactory, c.Workspaces)
	c.informerStopChan = make(chan struct{})
	err = controller.Run(c.informerStopChan)
	if err != nil {
		return log.DebugPrint(err)
	}
	c.informerController = controller
	c.informerStart = true

	return nil
}

type clusterController struct {
	//根据组_工作区记录集群
	//注意每次更新clusters,都需要更新cis表
	cis map[string]*Cluster //key:"group_workspace", Cluster
	//根据集群名记录集群
	clusters map[string]*Cluster //key: clusterName
	locker   sync.Mutex
}

func (c *clusterController) CreateOrUpdateCluster(group, workspace string, startinformer bool) (*Cluster, bool, error) {

	c.locker.Lock()
	defer c.locker.Unlock()

	gwkey := group + "_" + workspace
	_, ok := c.cis[gwkey]
	if ok {
		return nil, false, ErrClusterExists
	}

	token := "1234567890987654321"
	cConfig, err := GetK8sClientConfig(group, workspace, token)
	if err != nil {
		return nil, false, log.DebugPrint(err)
	}

	rconfig, err := ClusterConfigToK8sClientConfig(*cConfig)
	if err != nil {
		return nil, false, log.DebugPrint(err)
	}

	cluster, ok := c.clusters[cConfig.ClusterName]
	if ok {
		log.DebugPrint("start to update cluster :%v, group:%v,workspace:%v", cConfig.ClusterName, group, workspace)
		cluster.Reference += 1
		cluster.Workspaces[workspace] = Workspace{Name: workspace, Group: group}
		//log.DebugPrint("cluster :%v Informer's workspace:%v", cluster.Name, cluster.informerController.Workspaces)
		//cluster.informerController.Workspaces[workspace] = Workspace{Name: workspace, Group: group}
		c.cis[gwkey] = cluster
		c.clusters[cConfig.ClusterName] = cluster
		return cluster, true, nil
	} else {
		log.DebugPrint("start to create cluster :%v", cConfig.ClusterName)
		var cluster Cluster
		cluster.Name = cConfig.ClusterName
		cluster.Workspaces = make(map[string]Workspace)
		cluster.informerStopChan = make(chan struct{})
		cluster.Reference = 1
		cluster.Config = rconfig
		cluster.Workspaces[workspace] = Workspace{Name: workspace, Group: group}
		cluster.healthStopChan = make(chan struct{})
		cluster.IllCaused = cluster.ill()
		/*
			if startinformer && cluster.IllCaused == nil {
				log.DebugPrint("start inform er ....")
				err := cluster.StartInformers()
				if err != nil {
					return nil, false,err
				}
			}
		*/
		//		go cluster.HandleHealthyEvent()

		c.clusters[cConfig.ClusterName] = &cluster
		c.cis[gwkey] = &cluster
		return &cluster, false, nil
	}

}

func (c *clusterController) closeClusterInformers(clusterName string) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	cluster, ok := c.clusters[clusterName]
	if !ok {
		return fmt.Errorf("cluster '%v' not found", clusterName)
	}
	go cluster.CloseInformers()
	cluster.informerController = nil
	c.clusters[clusterName] = cluster
	for _, v := range cluster.Workspaces {
		key := v.Group + "_" + v.Name
		c.cis[key] = cluster
	}
	return nil
}

func (c *clusterController) startClusterInformers(clusterName string) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	cluster, ok := c.clusters[clusterName]
	if !ok {
		return fmt.Errorf("cluster '%v' not found", clusterName)
	}

	err := cluster.StartInformers()
	if err != nil {
		return err
	}

	go cluster.HandleHealthyEvent()

	c.clusters[clusterName] = cluster
	for _, v := range cluster.Workspaces {
		key := v.Group + "_" + v.Name
		c.cis[key] = cluster
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

	if !ci.informerStart || ci.IllCaused != nil {
		return nil, fmt.Errorf("cluster %v doesn't sync resource: '%v'", ci.Name, ci.IllCaused)
	}

	return ci, nil
}

func (c *clusterController) GetClusterIP(group, workspace string) (*string, error) {
	cluster, err := c.GetCluster(group, workspace)
	if err != nil {
		return nil, err
	}

	if cluster.Config != nil {
		return &cluster.Config.Host, nil
	}

	return nil, fmt.Errorf("cannot find cluster ip")
}

func (c *clusterController) GetClusterVIP(group, workspace string) (*string, error) {
	cluster, err := c.GetCluster(group, workspace)
	if err != nil {
		return nil, err
	}

	if cluster.Config != nil {
		u := cluster.Config.Host
		u = strings.TrimPrefix(u, "http://")
		u = strings.TrimPrefix(u, "https://")

		s := strings.Split(u, ":")
		host := s[0]
		return &host, nil
	}

	return nil, fmt.Errorf("cannot find cluster ip")
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
	cluster.locker.Lock()
	defer cluster.locker.Unlock()
	cluster.Reference -= 1
	//更新informers的工作区表,忽略指定工作区
	delete(cluster.Workspaces, workspace)

	//已经没有工作区了
	//停止cluster,即使后续会往集群中添加新的工作区
	if cluster.Reference == 0 {
		log.DebugPrint("start to delete cluster :%v", cluster.Name)
		cluster.CloseInformers()
		delete(c.clusters, pcluster.Name)
		go func() {
			cluster.healthStopChan <- struct{}{}
		}()
		//仍有工作区,仅仅清除组/工作区信息
	} else {
		log.DebugPrint("start to remove workspace in cluster : cluster '%v', group '%v', workspace,'%v/", pcluster.Name, group, workspace)
		c.clusters[pcluster.Name] = cluster
	}

	return nil
}

func (c *Cluster) HandleHealthyEvent() {
	defer log.DebugPrint("cluster '%v' healthy event checker exit....", c.Name)
	for {
		select {
		case <-c.healthStopChan:
			log.DebugPrint("recieve stop request...")
			return
		default:
		}

		caused := c.ill()
		/*
			if caused != nil {
				log.DebugPrint("cluster '%v' is ill:%v", c.Name, caused)
			}
		*/
		c.locker.Lock()
		switch {
		//之前出了问题,现在又正常了
		case caused == nil:
			//	log.DebugPrint("cluster '%v' is healthy again ,start to sync resource", c.Name)
			err := c.StartInformers()
			if err != nil {
				err = fmt.Errorf("cluster %v start informers fail for %v", err)
				c.IllCaused = err
			} else {

				c.IllCaused = nil
				//		log.DebugPrint("cluster '%v' is healthy", c.Name)
			}
			//之前健康,现在出了问题
		case caused != nil:
			c.CloseInformers()
			c.IllCaused = caused
		}
		c.locker.Unlock()

		//重新获取一次,避免此时cluster已经被删除

		time.Sleep(3 * time.Second)

	}
}

type Cluster struct {
	Name             string `json:"name"`
	Workspaces       map[string]Workspace
	informerStopChan chan struct{} //集群上各资源的informer stop chan
	Config           *rest.Config  //TODO:如何在证书修改时更新证书
	locker           sync.Mutex
	Reference        int //根据引用计数是否为1,可以判定CreateOrUpdateCluster是更新集群还是创建集群

	informerController *ResourceController
	clientset          *kubernetes.Clientset
	IllCaused          error
	informerStart      bool
	healthStopChan     chan struct{}
}
