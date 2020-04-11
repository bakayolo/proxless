package downscaler

import (
	"fmt"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes"
	"kube-proxless/internal/store/inmemory"
	"strconv"
	"time"
)

func StartDownScaler() {
	log.Info().Msgf("Starting DownScaler")
	clientDeployment := kubernetes.ClientSet.AppsV1().Deployments(config.Namespace)
	labelSelector := getProxyLabelSelector()

	for true { // infinite loop
		deploys, err := clientDeployment.List(labelSelector)
		if err != nil {
			log.Error().Err(err).Msgf(
				"Could not list deployments with label %s in namespace %s",
				labelSelector.LabelSelector, config.Namespace)
			// don't do anything else, we don't wanna kill the proxy
		} else {
			for _, deploy := range deploys.Items {
				if route, err := inmemory.GetRouteByDeploymentKey(deploy.Name, deploy.Namespace); err != nil {
					log.Error().Err(err).Msgf("Could not get route from map for deployment %s.%s", deploy.Name, config.Namespace)
				} else if *deploy.Spec.Replicas > int32(0) {
					timeIdle := time.Now().Sub(route.GetLastUsed()).Seconds()
					if timeIdle >= float64(config.ServerlessTTL) {
						scaleDownDeployment(deploy, clientDeployment)
					}
				}
			}
		}
		time.Sleep(time.Duration(config.ServerlessPollInterval) * time.Second)
	}
}

func getProxyLabelSelector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", kubernetes.LabelProxless, "true"),
	}
}

func scaleDownDeployment(deploy v1.Deployment, clientDeployment v1client.DeploymentInterface) {
	deploy.Spec.Replicas = int32Ptr(0)
	if _, err := clientDeployment.Update(&deploy); err != nil {
		log.Error().Err(err).Msgf("Could not scale down the deployment %s.%s", deploy.Name, deploy.Namespace)
	} else {
		log.Debug().Msgf("Deployment %s.%s scaled down after %s secs", deploy.Name, deploy.Namespace, strconv.Itoa(config.ServerlessTTL))
	}
}

func int32Ptr(i int32) *int32 { return &i }
