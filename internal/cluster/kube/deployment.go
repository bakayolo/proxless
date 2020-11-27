package kube

import (
	"context"
	"encoding/json"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"kube-proxless/internal/logger"
	"time"
)

type patchInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int32  `json:"value"`
}

func getDeployment(clientSet kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	return clientSet.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func patchDeploymentReplicas(
	clientSet kubernetes.Interface, name, namespace string, replicas int) (*appsv1.Deployment, error) {

	payloadBytes, err := json.Marshal([]patchInt32Value{{
		Op:    "replace",
		Path:  "/spec/replicas",
		Value: int32(replicas),
	}})

	if err != nil {
		return nil, err
	}

	return clientSet.AppsV1().Deployments(namespace).Patch(
		context.TODO(), name, k8stypes.JSONPatchType, payloadBytes, metav1.PatchOptions{})
}

func scaleUpDeployment(clientSet kubernetes.Interface, name, namespace string, timeout int) error {
	_, err := patchDeploymentReplicas(clientSet, name, namespace, 1)

	if err != nil {
		logger.Errorf(err, "Could not scale up the deployment %s.%s", name, namespace)
		return err
	}

	return waitForDeploymentAvailable(clientSet, name, namespace, timeout)
}

func waitForDeploymentAvailable(clientSet kubernetes.Interface, name, namespace string, timeout int) error {
	now := time.Now()
	intervalSeconds := 5 * time.Second // TODO make this configurable
	err := wait.PollImmediate(intervalSeconds, time.Duration(timeout)*time.Second, func() (bool, error) {
		if deploy, err := getDeployment(clientSet, name, namespace); err != nil {
			logger.Errorf(err, "Could not get the deployment %s.%s", name, namespace)
			return true, err
		} else {
			if deploy.Status.AvailableReplicas >= 1 { // TODO make this configurable
				logger.Debugf("Deployment %s.%s scaled up successfully after %s",
					name, namespace, time.Now().Sub(now))
				return true, nil
			} else {
				logger.Debugf("Deployment %s.%s not ready yet", name, namespace)
			}
			return false, nil
		}
	})
	return err
}

func scaleDownDeployment(kubeClient kubernetes.Interface, deploymentName, namespace string) error {
	_, err := patchDeploymentReplicas(kubeClient, deploymentName, namespace, 0)

	if err != nil {
		logger.Errorf(err, "Could not scale down deployment %s.%s", deploymentName, namespace)
		return err
	} else {
		logger.Debugf("Deployment %s.%s scaled down", deploymentName, namespace)
	}

	return nil
}
