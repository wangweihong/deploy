package controllers

import (
	"fmt"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type VolumeAndVolumeMounts struct {
	Volume      corev1.Volume      `json:"volume"`
	VolumeMount corev1.VolumeMount `json:"volumemount"`
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
	if !volumeFound {
		err := fmt.Errorf("volume not found")
		return corev1.PodSpec{}, err
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
	var newPodSpec corev1.PodSpec

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
	var newPodSpec corev1.PodSpec

	newPodSpec.Volumes = make([]corev1.Volume, 0)
	for _, v := range podSpec.Volumes {
		var found bool
		for _, j := range newVolumes {
			if v.Name == j.Name {
				found = true
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
	var newPodSpec corev1.PodSpec

	newPodSpec.Volumes = make([]corev1.Volume, 0)
	var found bool
	var k int
	for k = range podSpec.Volumes {

		if podSpec.Volumes[k].Name == volumeName {
			found = true
		}
	}

	if !found {
		return newPodSpec, fmt.Errorf("volume '%v' not found in old podspec", volumeName)
	}

	newPodSpec.Volumes = append(newPodSpec.Volumes, podSpec.Volumes[:k-1]...)
	newPodSpec.Volumes = append(newPodSpec.Volumes, podSpec.Volumes[k+1:]...)

	return newPodSpec, nil
}
