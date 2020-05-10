package kube

import (
	"errors"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/utils"
	"time"
)

func getDeployment(clientSet kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	return clientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func updateDeployment(
	clientSet kubernetes.Interface, deploy *appsv1.Deployment, namespace string) (*appsv1.Deployment, error) {
	return clientSet.AppsV1().Deployments(namespace).Update(deploy)
}

func listDeployments(
	clientSet kubernetes.Interface, namespace string, options metav1.ListOptions) ([]appsv1.Deployment, error) {
	deploys, err := clientSet.AppsV1().Deployments(namespace).List(options)

	if err != nil {
		return nil, err
	}

	if len(deploys.Items) == 0 {
		return nil, errors.New(
			fmt.Sprintf("0 deployment found with label %s in namespace %s", options.LabelSelector, namespace))
	}

	return deploys.Items, nil
}

func scaleUpDeployment(clientSet kubernetes.Interface, name, namespace string, timeout int) error {
	deploy, err := getDeployment(clientSet, name, namespace)
	if err != nil {
		logger.Errorf(err, "Could not get the deployment %s.%s", name, namespace)
		return err
	}

	if *deploy.Spec.Replicas == 0 {
		deploy.Spec.Replicas = pointer.Int32Ptr(1) // TODO make this configurable

		if _, err := updateDeployment(clientSet, deploy, namespace); err != nil {
			logger.Errorf(err, "Could not scale up the deployment %s.%s", name, namespace)
			return err
		} else {
			return waitForDeploymentAvailable(clientSet, name, namespace, timeout)
		}
	} else {
		return nil
	}
}

func waitForDeploymentAvailable(clientSet kubernetes.Interface, name, namespace string, timeout int) error {
	now := time.Now()
	err := wait.PollImmediate(time.Second, time.Duration(timeout)*time.Second, func() (bool, error) {
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
		LabelSelector: fmt.Sprintf("%s=%s", cluster.LabelDeploymentProxless, "true"),
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
					deploy.Spec.Replicas = pointer.Int32Ptr(0)

					_, err := updateDeployment(kubeClient, &deploy, deploy.Namespace)
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
	labels := map[string]string{cluster.LabelDeploymentProxless: "true"}

	deploy, err := getDeployment(clientSet, name, namespace)

	if err != nil {
		return nil, err
	}

	if deploy.Labels == nil {
		deploy.Labels = map[string]string{}
	}
	deploy.Labels = utils.MergeMap(deploy.Labels, labels)

	return updateDeployment(clientSet, deploy, namespace)
}

func removeDeploymentLabel(clientSet kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	deploy, err := getDeployment(clientSet, name, namespace)

	if err != nil {
		return nil, err
	}

	if deploy.Labels != nil {
		delete(deploy.Labels, cluster.LabelDeploymentProxless)
	}

	return updateDeployment(clientSet, deploy, namespace)
}
