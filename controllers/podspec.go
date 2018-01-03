package controllers

import (
	"fmt"
	"ufleet-deploy/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

var (
	VolumeEmptyDir  = "emptydir"
	VolumeHostPath  = "hostpath"
	VolumePVC       = "pvc"
	VolumeConfigMap = "configmap"
	VolumeSecret    = "secret"
)

type VolumeMount struct {
	Name      string `json:"name"`
	ReadOnly  bool   `json:"readonly"`
	MountPath string `json:"mountpath"`
}

type ContainerVolumeMount struct {
	Name   string               `json:"name"` //容器名
	Mounts []corev1.VolumeMount `json:"mounts"`
}

type Volume struct {
	Type      string                                   `json:"type"`
	Name      string                                   `json:"name"`
	EmptyDir  corev1.EmptyDirVolumeSource              `json:"emptydir"`
	HostPath  corev1.HostPathVolumeSource              `json:"hostpath"`
	PVC       corev1.PersistentVolumeClaimVolumeSource `json:"pvc"`
	ConfigMap corev1.ConfigMapVolumeSource             `json:"configmap"`
	Secret    corev1.SecretVolumeSource                `json:"secret"`
}

type VolumeAndVolumeMounts struct {
	//	Volume      corev1.Volume      `json:"volume"`
	//	VolumeMount corev1.VolumeMount `json:"volumemount"`
	Volume  Volume                 `json:"volume"`
	CMounts []ContainerVolumeMount `json:"cmounts"`
}

func updatePodSpecContainerEnv(podSpec corev1.PodSpec, container string, envVar []corev1.EnvVar) (corev1.PodSpec, error) {
	var containerFound bool
	var containerIndex int

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerIndex = k
			containerFound = true
			//		podSpec.Containers[k].Env = envVar

			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}
	newPodSpec.Containers[containerIndex].Env = make([]corev1.EnvVar, 0)
	newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, envVar...)

	return newPodSpec, nil
}

func deletePodSpecContainerEnv(podSpec corev1.PodSpec, container string, env string) (corev1.PodSpec, error) {
	var containerFound bool
	var envFound bool
	var containerIndex int
	var envIndex int

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerIndex = k
			containerFound = true
			for i, j := range v.Env {
				if j.Name == env {
					envFound = true
					envIndex = i
					break
				}
			}
			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}
	if !envFound {
		err := fmt.Errorf("env not found")
		return corev1.PodSpec{}, err
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}
	newPodSpec.Containers[containerIndex].Env = make([]corev1.EnvVar, 0)

	newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, podSpec.Containers[containerIndex].Env[:envIndex]...)
	newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, podSpec.Containers[containerIndex].Env[envIndex+1:]...)
	return newPodSpec, nil

}

func addPodSpecContainerEnv(podSpec corev1.PodSpec, container string, envVar []corev1.EnvVar) (corev1.PodSpec, error) {
	var containerFound bool
	var containerIndex int

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerFound = true
			containerIndex = k
			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}

	newPodSpec.Containers[containerIndex].Env = make([]corev1.EnvVar, 0)

	newPodSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, podSpec.Containers[containerIndex].Env...)
	newPodSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar...)

	return newPodSpec, nil
}

func updatePodSpecContainerVolume(podSpec corev1.PodSpec, container string, volumeVar []corev1.VolumeMount) (corev1.PodSpec, error) {
	var containerFound bool
	var containerIndex int

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerIndex = k
			containerFound = true
			//		podSpec.Containers[k].Volume = volumeVar

			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}
	newPodSpec.Containers[containerIndex].VolumeMounts = make([]corev1.VolumeMount, 0)
	//newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[ContainerIndex].VolumeMounts, v)
	newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[containerIndex].VolumeMounts, volumeVar...)
	return newPodSpec, nil

}

func deletePodSpecContainerVolume(podSpec corev1.PodSpec, container string, volume string) (corev1.PodSpec, error) {
	var containerFound bool
	var volumeFound bool
	var containerIndex int
	var volumeIndex int

	for k, v := range podSpec.Containers {
		if v.Name == container {
			containerIndex = k
			containerFound = true
			for i, j := range v.VolumeMounts {
				if j.Name == volume {
					volumeFound = true
					volumeIndex = i
					break
				}
			}
			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}
	//没有引用volume,直接返回
	if !volumeFound {
		log.DebugPrint("not found volume '%v' mount in containers", volume)
		return podSpec, nil
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}
	newPodSpec.Containers[containerIndex].VolumeMounts = make([]corev1.VolumeMount, 0)

	newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[containerIndex].VolumeMounts, podSpec.Containers[containerIndex].VolumeMounts[:volumeIndex]...)
	newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[containerIndex].VolumeMounts, podSpec.Containers[containerIndex].VolumeMounts[volumeIndex+1:]...)
	return newPodSpec, nil

}

func addPodSpecContainerVolume(podSpec corev1.PodSpec, container string, volumeVar []corev1.VolumeMount) (corev1.PodSpec, error) {
	var containerFound bool
	var containerIndex int

	for k, v := range podSpec.Containers {
		log.DebugPrint(v.Name)
		if v.Name == container {
			containerFound = true
			containerIndex = k
			break
		}
	}

	if !containerFound {
		err := fmt.Errorf("container not found")
		return corev1.PodSpec{}, err
	}

	var newPodSpec corev1.PodSpec
	newPodSpec = podSpec
	newPodSpec.Containers = make([]corev1.Container, 0)

	for _, v := range podSpec.Containers {
		newPodSpec.Containers = append(newPodSpec.Containers, v)
	}

	newPodSpec.Containers[containerIndex].VolumeMounts = make([]corev1.VolumeMount, 0)

	newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[containerIndex].VolumeMounts, podSpec.Containers[containerIndex].VolumeMounts...)
	newPodSpec.Containers[containerIndex].VolumeMounts = append(newPodSpec.Containers[containerIndex].VolumeMounts, volumeVar...)

	return newPodSpec, nil
}

func addPodSpecVolume(podSpec corev1.PodSpec, newVolumes []corev1.Volume) (corev1.PodSpec, error) {
	newPodSpec := podSpec

	newPodSpec.Volumes = make([]corev1.Volume, 0)
	for _, v := range podSpec.Volumes {
		for _, j := range newVolumes {
			if v.Name == j.Name {
				return newPodSpec, fmt.Errorf("volume '%v' has exist in pod spec", j.Name)
			}
		}
	}
	newPodSpec.Volumes = append(newPodSpec.Volumes, podSpec.Volumes...)
	newPodSpec.Volumes = append(newPodSpec.Volumes, newVolumes...)

	return newPodSpec, nil
}

func updatePodSpecVolume(podSpec corev1.PodSpec, newVolumes []corev1.Volume) (corev1.PodSpec, error) {
	newPodSpec := podSpec

	newPodSpec.Volumes = make([]corev1.Volume, 0)
	for _, v := range podSpec.Volumes {
		var found bool
		for _, j := range newVolumes {
			if v.Name == j.Name {
				found = true
				break
			}
		}
		if !found {
			return newPodSpec, fmt.Errorf("volume '%v' not found in new podspec", v.Name)
		}
	}
	newPodSpec.Volumes = append(newPodSpec.Volumes, newVolumes...)

	return newPodSpec, nil
}

func deletePodSpecVolume(podSpec corev1.PodSpec, volumeName string) (corev1.PodSpec, error) {
	newPodSpec := podSpec

	newPodSpec.Volumes = make([]corev1.Volume, 0)
	var found bool
	var k int
	for k = range podSpec.Volumes {

		if podSpec.Volumes[k].Name == volumeName {
			found = true
			break
		}
	}

	if !found {
		return newPodSpec, fmt.Errorf("volume '%v' not found in old podspec", volumeName)
	}

	newPodSpec.Volumes = append(newPodSpec.Volumes, podSpec.Volumes[:k]...)
	newPodSpec.Volumes = append(newPodSpec.Volumes, podSpec.Volumes[k+1:]...)

	return newPodSpec, nil
}

func getSpecVolume(podSpec corev1.PodSpec) []Volume {
	vols := make([]Volume, 0)
	for _, v := range podSpec.Volumes {
		var vol Volume
		vol.Name = v.Name
		switch {
		case v.EmptyDir != nil:
			vol.Type = VolumeEmptyDir
			vol.EmptyDir = *v.EmptyDir
		case v.HostPath != nil:
			vol.Type = VolumeHostPath
			vol.HostPath = *v.HostPath
		case v.PersistentVolumeClaim != nil:
			vol.Type = VolumePVC
			vol.PVC = *v.PersistentVolumeClaim
		case v.ConfigMap != nil:
			vol.Type = VolumeConfigMap
			vol.ConfigMap = *v.ConfigMap
		case v.Secret != nil:
			vol.Type = VolumeSecret
			vol.Secret = *v.Secret
		default:
			continue
		}
		vols = append(vols, vol)
	}
	return vols
}

func getSpecVolumeAndVolumeMounts(podSpec corev1.PodSpec) []VolumeAndVolumeMounts {
	volumes := getSpecVolume(podSpec)
	vavms := make([]VolumeAndVolumeMounts, 0)
	for k, v := range volumes {
		vms := make([]ContainerVolumeMount, 0)

		for _, c := range podSpec.Containers {
			var cvm ContainerVolumeMount
			cvm.Mounts = make([]corev1.VolumeMount, 0)
			cvm.Name = c.Name
			var found bool
			for i, j := range c.VolumeMounts {
				if j.Name == v.Name {
					cvm.Mounts = append(cvm.Mounts, c.VolumeMounts[i])
					found = true
				}
			}
			if found {
				vms = append(vms, cvm)
			}
		}
		var vavm VolumeAndVolumeMounts
		vavm.Volume = volumes[k]
		vavm.CMounts = vms
		vavms = append(vavms, vavm)
	}
	return vavms
}

func addVolumeAndContaienrVolumeMounts(podSpec corev1.PodSpec, volumeVar VolumeAndVolumeMounts) (corev1.PodSpec, error) {

	//	var newPodSpec corev1.PodSpec
	//	newPodSpec = podSpec
	var err error
	var vol corev1.Volume
	vol.Name = volumeVar.Volume.Name
	switch volumeVar.Volume.Type {

	case VolumeEmptyDir:
		vol.EmptyDir = &volumeVar.Volume.EmptyDir
	case VolumeHostPath:
		vol.HostPath = &volumeVar.Volume.HostPath
	case VolumePVC:
		vol.PersistentVolumeClaim = &volumeVar.Volume.PVC
	case VolumeConfigMap:
		vol.ConfigMap = &volumeVar.Volume.ConfigMap
	case VolumeSecret:
		vol.Secret = &volumeVar.Volume.Secret
	default:
		err = fmt.Errorf("unsupported volume type: %v", volumeVar.Volume.Type)
		return corev1.PodSpec{}, err
	}

	newPodSpec, err := addPodSpecVolume(podSpec, []corev1.Volume{vol})
	if err != nil {
		return newPodSpec, err
	}

	for _, v := range volumeVar.CMounts {
		newPodSpec, err = addPodSpecContainerVolume(newPodSpec, v.Name, v.Mounts)
		if err != nil {
			return newPodSpec, err
		}
	}

	return newPodSpec, nil
}

func deleteVolumeAndContaienrVolumeMounts(podSpec corev1.PodSpec, volume string) (corev1.PodSpec, error) {

	newPodSpec := podSpec
	var err error
	newPodSpec, err = deletePodSpecVolume(podSpec, volume)
	if err != nil {
		return corev1.PodSpec{}, log.DebugPrint(err)
	}

	for _, v := range newPodSpec.Containers {
		newPodSpec, err = deletePodSpecContainerVolume(newPodSpec, v.Name, volume)
		if err != nil {
			return corev1.PodSpec{}, log.DebugPrint(err)
		}
	}

	return newPodSpec, nil
}
