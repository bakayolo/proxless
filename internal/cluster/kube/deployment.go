package kube

import (
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeDeploymentInterface interface {
	getDeployment(name, namespace string) (*v1.Deployment, error)
	updateDeployment(deploy *v1.Deployment, namespace string) (*v1.Deployment, error)
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
