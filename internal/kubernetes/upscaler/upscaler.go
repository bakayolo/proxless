package upscaler

import (
	"fmt"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes"
	"time"
)

func ScaleUpDeployment(name, namespace string) error {
	clientDeployment := kubernetes.ClientSet.AppsV1().Deployments(namespace)
	deploy, err := clientDeployment.Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Could not get deployment %s.%s", name, namespace)
		return err
	}

	return scaleUp(*deploy, clientDeployment)
}

func scaleUp(deploy v1.Deployment, clientDeployment v1client.DeploymentInterface) error {
	deploy.Spec.Replicas = pointer.Int32Ptr(1)
	if _, err := clientDeployment.Update(&deploy); err != nil {
		log.Error().Err(err).Msgf("Could not scale up the deployment %s.%s", deploy.Name, deploy.Namespace)
		return err
	} else {
		return waitForDeploymentAvailable(deploy, clientDeployment)
	}
}

func waitForDeploymentAvailable(deploy v1.Deployment, clientDeployment v1client.DeploymentInterface) error {
	// TODO understand the Interval value
	pollInterval, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollInterval))
	pollTimeout, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollTimeout))
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		if deploy, err := clientDeployment.Get(deploy.Name, metav1.GetOptions{}); err != nil {
			log.Error().Err(err).Msgf("Could not get the deployment %s.%s", deploy.Name, deploy.Namespace)
			return true, err
		} else {
			if deploy.Status.AvailableReplicas >= 1 { // TODO make this configurable
				log.Debug().Msgf("Deployment %s.%s scaled up successfully", deploy.Name, deploy.Namespace)
				return true, nil
			} else {
				log.Debug().Msgf("Deployment %s.%s not ready yet", deploy.Name, deploy.Namespace)
			}
			return false, nil
		}
	})
	return err
}
