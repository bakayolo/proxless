package kube

import (
	"errors"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/cluster"
	"strings"
)

type mockDeploymentClient struct{}

func (m *mockDeploymentClient) getDeployment(name, namespace string) (*v1.Deployment, error) {
	if name == "err" {
		return nil, errors.New("deployment not found")
	}

	availability := 1
	if name == "timeout" {
		availability = 0
	}

	return &v1.Deployment{
		Status: v1.DeploymentStatus{
			AvailableReplicas: int32(availability),
		},
	}, nil
}

func (m *mockDeploymentClient) updateDeployment(deploy *v1.Deployment, namespace string) (*v1.Deployment, error) {
	if deploy.Name == "err" {
		return nil, errors.New("deployment not found")
	}

	return &v1.Deployment{}, nil
}

func (m *mockDeploymentClient) listDeployments(namespace string, options metav1.ListOptions) ([]v1.Deployment, error) {
	if strings.Contains(options.LabelSelector, cluster.LabelDeploymentProxless) {
		return []v1.Deployment{
			{
				Spec: v1.DeploymentSpec{Replicas: pointer.Int32Ptr(1)},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "err"},
				Spec:       v1.DeploymentSpec{Replicas: pointer.Int32Ptr(1)},
			},
		}, nil
	}

	return nil, errors.New("deployments not found")
}

func (m *mockDeploymentClient) labelDeployment(name, namespace string, labels map[string]string) (*v1.Deployment, error) {
	return nil, nil
}

func (m *mockDeploymentClient) unlabelDeployment(name, namespace string, label string) (*v1.Deployment, error) {
	return nil, nil
}
