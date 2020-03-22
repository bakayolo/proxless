package kubernetes

import (
	"fmt"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"kube-proxless/internal/config"
	"kube-proxless/internal/store"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup
)

func ScaleUp(labelValue, namespace string) error {
	clientDeployment := clientSet.AppsV1().Deployments(namespace)
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", config.LabelProxlessSvc, labelValue),
	}
	deploys, err := clientDeployment.List(labelSelector)
	if err != nil {
		log.Error().Err(err).Msgf("Could not list deployments with label %s=%s in namespace %s", config.LabelProxlessSvc, labelValue, namespace)
		return err
	} else if len(deploys.Items) > 0 {
		deployLen := len(deploys.Items)
		if deployLen > 1 {
			log.Error().Err(err).Msgf("More than one deployment with label %s=%s in namespace %s", config.LabelProxlessSvc, labelValue, namespace)
			// should not have more than 1 deployment corresponding to the label
			// however we are still gonna scale up all of the deployments to not make the gateway fail
		}

		ch := make(chan error, deployLen)
		defer close(ch)
		wg.Add(deployLen)

		for _, item := range deploys.Items {
			go scaleUpDeployment(item, clientDeployment, ch)
		}
		wg.Wait()
	}
	return nil
}

func scaleUpDeployment(item v1.Deployment, clientDeployment v1client.DeploymentInterface, ch chan error) {
	item.Spec.Replicas = int32Ptr(1)
	if _, err := clientDeployment.Update(&item); err != nil {
		log.Error().Err(err).Msgf("Could not scale up the deployment %s in namespace %s", item.Name, item.Namespace)
		ch <- err
	} else {
		// TODO understand the Interval value
		pollInterval, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollInterval))
		pollTimeout, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollTimeout))
		err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
			if deploy, err := clientDeployment.Get(item.Name, metav1.GetOptions{}); err != nil {
				log.Error().Err(err).Msgf("Could not get the deployment %s in namespace %s", item.Name, item.Namespace)
				return true, err
			} else {
				if deploy.Status.AvailableReplicas >= 1 {
					log.Debug().Msgf("Deployment %s scaled up successfully in namespace %s", item.Name, item.Namespace)
					return true, nil
				} else {
					log.Debug().Msgf("Deployment %s still rolling in namespace %s", item.Name, item.Namespace)
				}
				return false, nil
			}
		})
		ch <- err
	}
	wg.Done()
}

func ScalingEngine() {
	log.Info().Msgf("Starting Scaling Engine")
	clientDeployment := clientSet.AppsV1().Deployments(config.Namespace)
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", config.LabelProxlessEnabled, "true"),
	}
	for true {
		deploys, err := clientDeployment.List(labelSelector)
		if err != nil {
			log.Error().Err(err).Msgf("Could not list deployments with label %s=true in namespace %s", config.LabelProxlessSvc, config.Namespace)
			// don't do anything else, we don't wanna kill the proxy
		} else {
			for _, item := range deploys.Items {
				label := item.Labels[config.LabelProxlessSvc]
				if route, err := store.GetRouteByLabel(label); err != nil {
					log.Error().Err(err).Msgf("Could not get route from map for deployment with label %s=true in namespace %s", config.LabelProxlessSvc, config.Namespace)
				} else if *item.Spec.Replicas > int32(0) {
					timeIdle := time.Now().Sub(route.LastUsed).Seconds()
					if timeIdle >= float64(config.ServerlessTTL) {
						item.Spec.Replicas = int32Ptr(0)
						if _, err := clientDeployment.Update(&item); err != nil {
							log.Error().Err(err).Msgf("Could not scale down the deployment %s in namespace %s", item.Name, item.Namespace)
						} else {
							log.Debug().Msgf("Deployment %s in namespace %s scaled down after %s", item.Name, item.Namespace, config.ServerlessTTL)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(config.ServerlessPollInterval) * time.Second)
	}
}

func int32Ptr(i int32) *int32 { return &i }
