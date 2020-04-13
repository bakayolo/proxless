package kube

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/config"
	"time"
)

type KubeClient struct {
	deployClient KubeDeploymentInterface
}

func NewKubeClient() *KubeClient {
	kubeConf := loadKubeConfig(config.KubeConfigPath)
	clientSet := kubernetes.NewForConfigOrDie(kubeConf)

	return &KubeClient{
		deployClient: &KubeDeploymentClient{
			clientSet: clientSet,
		},
	}
}

func loadKubeConfig(kubeConfigPath string) *rest.Config {
	kubeConfigString := flag.String("kubeconfig", kubeConfigPath, "(optional) absolute path to the kubeconfig file")

	// use the current context in kubeconfig
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeConfigString)
	if err != nil {
		log.Panic().Err(err).Msgf("Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return kubeConfig
}

func (c *KubeClient) ScaleUpDeployment(name, namespace string) error {
	deploy, err := c.deployClient.getDeployment(name, namespace)
	if err != nil {
		log.Error().Err(err).Msgf("Could not get deployment %s.%s", name, namespace)
		return err
	}

	deploy.Spec.Replicas = pointer.Int32Ptr(1)
	if _, err := c.deployClient.updateDeployment(deploy, namespace); err != nil {
		log.Error().Err(err).Msgf("Could not scale up the deployment %s.%s", name, namespace)
		return err
	} else {
		return c.waitForDeploymentAvailable(name, namespace)
	}
}

func (c *KubeClient) waitForDeploymentAvailable(name, namespace string) error {
	pollInterval, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollInterval))
	pollTimeout, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollTimeout))
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		if deploy, err := c.deployClient.getDeployment(name, namespace); err != nil {
			log.Error().Err(err).Msgf("Could not get the deployment %s.%s", name, namespace)
			return true, err
		} else {
			if deploy.Status.AvailableReplicas >= 1 { // TODO make this configurable
				log.Debug().Msgf("Deployment %s.%s scaled up successfully", name, namespace)
				return true, nil
			} else {
				log.Debug().Msgf("Deployment %s.%s not ready yet", name, namespace)
			}
			return false, nil
		}
	})
	return err
}
