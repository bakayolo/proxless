package kube

import (
	"errors"
	v1 "k8s.io/api/apps/v1"
)

type mockDeploymentClient struct{}

func (m *mockDeploymentClient) getDeployment(name, namespace string) (*v1.Deployment, error) {
	if name == "err" {
		return nil, errors.New("Deployment not found")
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
		return nil, errors.New("Deployment not found")
	}

	return &v1.Deployment{}, nil
}
