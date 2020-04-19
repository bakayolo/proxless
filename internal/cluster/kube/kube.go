package kube

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/cluster"
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

func (c *KubeClient) RunDownScaler(
	namespace string,
	checkInterval time.Duration,
	mustScaleDown func(deployName, namespace string) bool,
) {
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", cluster.LabelDeploymentProxless, "true"),
	}

	for {
		c.scaleDown(namespace, labelSelector, mustScaleDown)

		time.Sleep(checkInterval)
	}
}

func (c *KubeClient) scaleDown(
	namespace string,
	labelSelector metav1.ListOptions,
	mustScaleDown func(deployName, namespace string) bool,
) {
	deploys, err := c.deployClient.listDeployment(namespace, labelSelector)

	if err != nil {
		log.Error().Err(err).Msgf(
			"Could not list deployments with label %s in namespace %s",
			labelSelector.LabelSelector, namespace)
		// don't do anything else, we don't wanna kill the proxy
	} else {
		for _, deploy := range deploys {
			if *deploy.Spec.Replicas > int32(0) && mustScaleDown(deploy.Name, deploy.Namespace) {
				deploy.Spec.Replicas = pointer.Int32Ptr(0)

				_, err := c.deployClient.updateDeployment(&deploy, deploy.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf(""+
						"Could not scale down the deployment %s.%s",
						deploy.Name, deploy.Namespace)
				}
			}
		}
	}
}
