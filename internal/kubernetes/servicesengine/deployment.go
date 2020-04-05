package servicesengine

import (
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/kubernetes"
)

func addProxyLabelToDeployment(name, namespace string) {
	err := addLabelToDeployment(name, namespace, kubernetes.LabelProxless)
	if err != nil {
		log.Error().Err(err).Msgf(
			"Cannot add label on deployment %s.%s", name, namespace)
	}
}

func addLabelToDeployment(name, namespace, label string) error {
	clientDeployment := kubernetes.ClientSet.AppsV1().Deployments(namespace)
	deploy, err := clientDeployment.Get(name, v1.GetOptions{})

	if err != nil {
		return err
	}

	deploy.Labels[label] = "true"
	_, err = clientDeployment.Update(deploy)

	return err
}

func removeProxyLabelFromDeployment(name, namespace string) {
	err := removeLabelFromDeployment(name, namespace, kubernetes.LabelProxless)
	if err != nil {
		log.Error().Err(err).Msgf(
			"Cannot remove label from deployment %s.%s", name, namespace)
	}
}

func removeLabelFromDeployment(name, namespace, label string) error {
	clientDeployment := kubernetes.ClientSet.AppsV1().Deployments(namespace)
	deploy, err := clientDeployment.Get(name, v1.GetOptions{})

	if err != nil {
		return err
	}

	if deploy.Labels[label] != "" {
		delete(deploy.Labels, label)
	}
	_, err = clientDeployment.Update(deploy)

	return err
}
