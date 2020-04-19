package kube

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeDeploymentInterface interface {
	getDeployment(name, namespace string) (*v1.Deployment, error)
	updateDeployment(deploy *v1.Deployment, namespace string) (*v1.Deployment, error)
	listDeployment(namespace string, options metav1.ListOptions) ([]v1.Deployment, error)
}

type KubeDeploymentClient struct {
	clientSet *kubernetes.Clientset
}

func (c *KubeDeploymentClient) getDeployment(name, namespace string) (*v1.Deployment, error) {
	return c.clientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func (c *KubeDeploymentClient) updateDeployment(deploy *v1.Deployment, namespace string) (*v1.Deployment, error) {
	return c.clientSet.AppsV1().Deployments(namespace).Update(deploy)
}

func (c *KubeDeploymentClient) listDeployment(namespace string, options metav1.ListOptions) ([]v1.Deployment, error) {
	deploys, err := c.clientSet.AppsV1().Deployments(namespace).List(options)

	if err != nil {
		return nil, err
	}

	if len(deploys.Items) == 0 {
		return nil, errors.New(fmt.Sprintf("0 deployment found with label %s in namespace %s", options.LabelSelector, namespace))
	}

	return deploys.Items, nil
}
