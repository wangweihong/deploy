package hpa

import (
	"fmt"
	"strings"
	"ufleet-deploy/pkg/backend"
	"ufleet-deploy/pkg/cluster"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/deployment"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

func HandleEventWatchFromK8sCluster(echan chan cluster.Event, kind string, oc *HorizontalPodAutoscalerManager) {

	//可以考虑ufleet创建的configmap直接绑定k8s configmap,
	//在k8s configmap更新的时候再更新绑定的k8s  configmap.
	//在删除的,标记底层资源已经被移除.
	log.DebugPrint(" %v cluster  event handler start !", strings.ToUpper(kind))
	defer log.ErrorPrint("%v cluster  event handler finish !", strings.ToUpper(kind))

	for {
		pe := <-echan

		go func(e cluster.Event) {
			switch e.Action {
			case cluster.ActionDelete:
				//清除内存中的数据即可
				obj, err := oc.GetObject(e.Group, e.Workspace, e.Name)
				if err != nil {
					if err == resource.ErrResourceNotFound {
						return
					}
					log.ErrorPrint("%v:  event handler delete 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
					return
				}

				//如果memoryOnly不为true,内存中的数据由etcd去清理
				if obj.Metadata().MemoryOnly {
					err := oc.DeleteObject(e.Group, e.Workspace, e.Name, resource.DeleteOption{MemoryOnly: true})
					if err != nil {
						if err != resource.ErrGroupNotFound || err != resource.ErrWorkspaceNotFound {
							log.ErrorPrint("%v:  event handler delete 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
						}
					}
				}

				go func() {
					err := oc.setDeploymentHPA(e)
					if err != nil {
						log.ErrorPrint(err)
					}
				}()
				return

			case cluster.ActionCreate:
				if e.FromUfleet {
					go func() {
						err := oc.setDeploymentHPA(e)
						if err != nil {
							log.ErrorPrint(err)
						}
					}()
					return
				}
				oc.Lock()
				defer oc.Unlock()

				_, err := oc.GetObjectWithoutLock(e.Group, e.Workspace, e.Name)
				if err != nil {
					if err != resource.ErrResourceNotFound {
						log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)
						return
					}
				} else {
					log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v': exists ", kind, e.Group, e.Workspace, e.Name)
					return
				}

				var p resource.ObjectMeta
				p.Name = e.Name
				p.MemoryOnly = true
				p.Workspace = e.Workspace
				p.Group = e.Group
				p.User = resource.ClusterObjectCreater
				p.Kind = kind

				err = oc.NewObject(p)
				if err != nil {
					if err != resource.ErrResourceExists {
						log.ErrorPrint("%v:  event handler create 'group:%v,Workspace:%v,resource:%v':%v ", kind, e.Group, e.Workspace, e.Name, err)

						return
					}
				}
				go func() {
					err := oc.setDeploymentHPA(e)
					if err != nil {
						log.ErrorPrint(err)
					}
				}()

				return
			case cluster.ActionUpdate:
				go func() {
					err := oc.setDeploymentHPA(e)
					if err != nil {
						log.ErrorPrint(err)
					}
				}()

				return
			}
		}(pe)

	}
}

func (c *HorizontalPodAutoscalerManager) setDeploymentHPA(e cluster.Event) error {
	hpaObj, _ := e.Object.(*autoscalingv1.HorizontalPodAutoscaler)
	hpaspec := hpaObj.Spec

	if hpaspec.ScaleTargetRef.Kind != "Deployment" {
		err := fmt.Errorf("hpa scale target is not deployment, not support")
		return err
	}

	d, err := pk.Controller.GetObject(e.Group, e.Workspace, hpaObj.Spec.ScaleTargetRef.Name)
	if err != nil {
		if err == resource.ErrResourceNotFound {
			return nil
		}
		return err
	}

	if hpaObj.Spec.TargetCPUUtilizationPercentage != nil {
		minCPU := int(float32(*hpaObj.Spec.TargetCPUUtilizationPercentage) * 0.8)
		maxCPU := int(float32(*hpaObj.Spec.TargetCPUUtilizationPercentage) * 1.2)

		var hpaopt pk.HPA
		switch e.Action {
		case cluster.ActionCreate, cluster.ActionUpdate:
			hpaopt.Type = "cpu"
			hpaopt.MinCPU = minCPU
			hpaopt.MaxCPU = maxCPU
			if hpaspec.MinReplicas != nil {
				hpaopt.MinReplicas = int(*hpaspec.MinReplicas)
			}
			hpaopt.MaxReplicas = int(hpaspec.MaxReplicas)

		case cluster.ActionDelete:
			hpaopt.Type = "none"
		}

		di, _ := pk.GetDeploymentInterface(d)
		err := di.StartAutoScale(hpaopt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *HorizontalPodAutoscalerManager) HandleEvent(e backend.ResourceEvent) {
	resource.EtcdEventHandler(e, c)
}
