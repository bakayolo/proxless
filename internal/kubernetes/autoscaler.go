package kubernetes

import (
	"fmt"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"kube-proxless/internal/config"
	"sync"
	"time"
)

func ScaleUp(labelValue, namespace string) error {
	clientDeployment := clientSet.AppsV1().Deployments(namespace)
	labelSelector := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelSvc, labelValue),
	}
	deploys, err := clientDeployment.List(labelSelector)
	if err != nil {
		log.Error().Err(err).Msgf("Could not list deployments with label %s=%s in namespace %s", labelSvc, labelValue, namespace)
		return err
	} else if len(deploys.Items) > 0 {
		deployLen := len(deploys.Items)
		if deployLen > 1 {
			log.Error().Err(err).Msgf("More than one deployment with label %s=%s in namespace %s", labelSvc, labelValue, namespace)
			// should not have more than 1 deployment corresponding to the label
			// however we are still gonna scale up all of the deployments to not make the gateway fail
		}

		ch := make(chan error, deployLen)
		defer close(ch)
		var wg sync.WaitGroup
		wg.Add(deployLen)

		for _, item := range deploys.Items {
			go func(item v1.Deployment) {
				item.Spec.Replicas = int32Ptr(1)

				if _, err := clientDeployment.Update(&item); err != nil {
					ch <- err
				} else {
					// TODO understand the Interval value
					pollInterval, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollInterval))
					pollTimeout, _ := time.ParseDuration(fmt.Sprintf("%ss", config.ReadinessPollTimeout))
					err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
						if deploy, err := clientDeployment.Get(item.ObjectMeta.Name, metav1.GetOptions{}); err != nil {
							log.Error().Err(err).Msgf("Could not get the deployment %s in namespace %s", item.ObjectMeta.Name, namespace)
							return true, err
						} else {
							//
							if deploy.Status.AvailableReplicas >= 1 {
								log.Debug().Msgf("Deployment %s scaled up successfully in namespace %s", item.ObjectMeta.Name, namespace)
								return true, nil
							} else {
								log.Debug().Msgf("Deployment %s still rolling in namespace %s", item.ObjectMeta.Name, namespace)
							}
							return false, nil
						}
					})
					ch <- err
				}
				wg.Done()
			}(item)
		}
		wg.Wait()
	}
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
