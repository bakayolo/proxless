package openshift

import (
	"context"
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/client-go/apps/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	utils2 "kube-proxless/internal/cluster/utils"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/utils"
	"time"
)

func getDeployment(clientSet versioned.Interface, name, namespace string) (*appsv1.DeploymentConfig, error) {
	return clientSet.AppsV1().DeploymentConfigs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func updateDeployment(
	clientSet versioned.Interface, deploy *appsv1.DeploymentConfig, namespace string) (*appsv1.DeploymentConfig, error) {
	return clientSet.AppsV1().DeploymentConfigs(namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
}

func listDeployments(
	clientSet versioned.Interface, namespace string, options metav1.ListOptions) ([]appsv1.DeploymentConfig, error) {

	deploys, err := clientSet.AppsV1().DeploymentConfigs(namespace).List(context.TODO(), options)

	if err != nil {
		return nil, err
	}

	return deploys.Items, nil
}

func scaleUpDeployment(clientSet versioned.Interface, name, namespace string, timeout int) error {
	deploy, err := getDeployment(clientSet, name, namespace)
	if err != nil {
		logger.Errorf(err, "Could not get the deployment %s.%s", name, namespace)
		return err
	}

	if deploy.Spec.Replicas == 0 {
		deploy.Spec.Replicas = 1 // TODO make this configurable

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

func waitForDeploymentAvailable(clientSet versioned.Interface, name, namespace string, timeout int) error {
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
	clientSet versioned.Interface,
	namespaceScope string,
	mustScaleDown func(deployName, namespace string) (bool, time.Duration, error),
) []error {
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", utils2.LabelDeploymentProxless, "true"),
	}

	deploys, err := listDeployments(clientSet, namespaceScope, labelSelector)

	var errs []error
	if err != nil {
		logger.Errorf(
			err,
			"Could not list deployments with label %s in namespace %s",
			labelSelector.LabelSelector, namespaceScope)
		errs = append(errs, err)
	} else {
		for _, deploy := range deploys {
			if deploy.Spec.Replicas > int32(0) {
				scaleDown, timeIdle, _ := mustScaleDown(deploy.Name, deploy.Namespace)
				if scaleDown {
					deploy.Spec.Replicas = 0

					_, err := updateDeployment(clientSet, &deploy, deploy.Namespace)
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

func labelDeployment(clientSet versioned.Interface, name, namespace string) (*appsv1.DeploymentConfig, error) {
	labels := map[string]string{utils2.LabelDeploymentProxless: "true"}

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

func removeDeploymentLabel(
	clientSet versioned.Interface, name, namespace string) (*appsv1.DeploymentConfig, error) {
	deploy, err := getDeployment(clientSet, name, namespace)

	if err != nil {
		return nil, err
	}

	if deploy.Labels != nil {
		delete(deploy.Labels, utils2.LabelDeploymentProxless)
	}

	return updateDeployment(clientSet, deploy, namespace)
}
