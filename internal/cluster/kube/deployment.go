package kube

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kube-proxless/internal/utils"
)

type KubeDeploymentInterface interface {
	getDeployment(name, namespace string) (*v1.Deployment, error)
	updateDeployment(deploy *v1.Deployment, namespace string) (*v1.Deployment, error)
	listDeployments(namespace string, options metav1.ListOptions) ([]v1.Deployment, error)
	labelDeployment(name, namespace string, labels map[string]string) (*v1.Deployment, error)
	unlabelDeployment(name, namespace string, label string) (*v1.Deployment, error)
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

func (c *KubeDeploymentClient) listDeployments(namespace string, options metav1.ListOptions) ([]v1.Deployment, error) {
	deploys, err := c.clientSet.AppsV1().Deployments(namespace).List(options)

	if err != nil {
		return nil, err
	}

	if len(deploys.Items) == 0 {
		return nil, errors.New(fmt.Sprintf("0 deployment found with label %s in namespace %s", options.LabelSelector, namespace))
	}

	return deploys.Items, nil
}

func (c *KubeDeploymentClient) labelDeployment(name, namespace string, labels map[string]string) (*v1.Deployment, error) {
	deploy, err := c.getDeployment(name, namespace)

	if err != nil {
		return nil, err
	}

	if deploy.Labels == nil {
		deploy.Labels = map[string]string{}
	}
	deploy.Labels = utils.MergeMap(deploy.Labels, labels)

	return c.updateDeployment(deploy, namespace)
}

func (c *KubeDeploymentClient) unlabelDeployment(name, namespace string, label string) (*v1.Deployment, error) {
	deploy, err := c.getDeployment(name, namespace)

	if err != nil {
		return nil, err
	}

	if deploy.Labels == nil {
		delete(deploy.Labels, label)
	}

	return c.updateDeployment(deploy, namespace)
}
