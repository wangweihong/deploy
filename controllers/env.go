package controllers

import (
	"fmt"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

func updatePodSpecEnv(podSpec corev1.PodSpec, container string, envVar []corev1.EnvVar) (corev1.PodSpec, error) {
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
	for _, v := range podSpec.Containers[containerIndex].Env {
		newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, v)
	}
	newPodSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar...)
	return newPodSpec, nil

}

func deletePodSpecEnv(podSpec corev1.PodSpec, container string, env string) (corev1.PodSpec, error) {
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
	for _, v := range podSpec.Containers[containerIndex].Env {
		newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, v)
	}

	newPodSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env[:envIndex], podSpec.Containers[containerIndex].Env[envIndex+1:]...)
	return newPodSpec, nil

}

func addPodSpecEnv(podSpec corev1.PodSpec, container string, envVar []corev1.EnvVar) (corev1.PodSpec, error) {
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
	for _, v := range podSpec.Containers[containerIndex].Env {
		newPodSpec.Containers[containerIndex].Env = append(newPodSpec.Containers[containerIndex].Env, v)
	}
	newPodSpec.Containers[containerIndex].Env = append(podSpec.Containers[containerIndex].Env, envVar...)

	return newPodSpec, nil
}
