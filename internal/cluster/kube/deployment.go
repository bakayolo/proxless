package kube

import (
	"context"
	"encoding/json"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clusterutils "kube-proxless/internal/cluster/utils"
	"kube-proxless/internal/logger"
	"time"
)

type patchInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int32  `json:"value"`
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
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

func listDeployments(
	clientSet kubernetes.Interface, namespace string, options metav1.ListOptions) ([]appsv1.Deployment, error) {
	deploys, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), options)

	if err != nil {
		return nil, err
	}

	return deploys.Items, nil
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

func scaleDownDeployments(
	kubeClient kubernetes.Interface,
	namespaceScope string,
	mustScaleDown func(deployName, namespace string) (bool, time.Duration, error),
) []error {
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", clusterutils.LabelDeploymentProxless, "true"),
	}

	deploys, err := listDeployments(kubeClient, namespaceScope, labelSelector)

	var errs []error
	if err != nil {
		logger.Errorf(
			err,
			"Could not list deployments with label %s in namespace %s",
			labelSelector.LabelSelector, namespaceScope)
		errs = append(errs, err)
	} else {
		for _, deploy := range deploys {
			if *deploy.Spec.Replicas > int32(0) {
				scaleDown, timeIdle, _ := mustScaleDown(deploy.Name, deploy.Namespace)
				if scaleDown {
					_, err := patchDeploymentReplicas(kubeClient, deploy.Name, deploy.Namespace, 0)

					if err != nil {
						logger.Errorf(err, "Could not scale down deployment %s.%s", deploy.Name, deploy.Namespace)
						errs = append(errs, err)
					} else {
						logger.Debugf("Deployment %s.%s scaled down after %s",
							deploy.Name, deploy.Namespace, timeIdle)
					}
				}
			}
		}
	}
	return errs
}

func labelDeployment(clientSet kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	payloadBytes, err := json.Marshal([]patchStringValue{{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/labels/%s", clusterutils.LabelDeploymentProxless),
		Value: "true",
	}})

	if err != nil {
		return nil, err
	}

	return clientSet.AppsV1().Deployments(namespace).Patch(
		context.TODO(), name, k8stypes.JSONPatchType, payloadBytes, metav1.PatchOptions{})
}

func removeDeploymentLabel(clientSet kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	payloadBytes, err := json.Marshal([]patchStringValue{{
		Op:   "remove",
		Path: fmt.Sprintf("/metadata/labels/%s", clusterutils.LabelDeploymentProxless),
	}})

	if err != nil {
		return nil, err
	}

	return clientSet.AppsV1().Deployments(namespace).Patch(
		context.TODO(), name, k8stypes.JSONPatchType, payloadBytes, metav1.PatchOptions{})
}
