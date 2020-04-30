package kube

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/utils"
	"testing"
)

func Test_scaleUpDeployment(t *testing.T) {
	TestClusterClient_ScaleUpDeployment(t)
}

func Test_waitForDeploymentAvailable(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	timeout := 1

	// error - deployment is not in kubernetes
	assert.Error(t, waitForDeploymentAvailable(clientSet, dummyProxlessName, dummyNamespaceName, timeout))

	helper_createNamespace(t, clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, clientSet)

	// error - deployment in kubernetes but not available
	assert.Error(t, waitForDeploymentAvailable(clientSet, dummyProxlessName, dummyNamespaceName, timeout))

	deploy.Status.AvailableReplicas = 1
	helper_updateDeployment(t, clientSet, deploy)

	// no error - deployment in kubernetes and available
	assert.NoError(t, waitForDeploymentAvailable(clientSet, dummyProxlessName, dummyNamespaceName, timeout))
}

func Test_scaleDownDeployments(t *testing.T) {
	TestClusterClient_ScaleDownDeployments(t)
}

func Test_labelDeployment(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	// error - deployment is not in kubernetes
	_, err := labelDeployment(clientSet, dummyNonProxlessName, dummyNamespaceName)
	assert.Error(t, err)

	helper_createNamespace(t, clientSet)
	deploy := helper_createRandomDeployment(t, clientSet)

	// no error - deployment in kubernetes
	deploy, err = labelDeployment(clientSet, dummyNonProxlessName, dummyNamespaceName)
	assert.NoError(t, err)

	wantLabels := map[string]string{cluster.LabelDeploymentProxless: "true"}
	if !utils.CompareMap(deploy.Labels, wantLabels) {
		t.Errorf("labelDeployment(); labels = %s, wantLabels = %s", deploy.Labels, wantLabels)
	}
}

func Test_removeDeploymentLabel(t *testing.T) {
	clientSet := fake.NewSimpleClientset()

	// error - deployment is not in kubernetes
	_, err := removeDeploymentLabel(clientSet, dummyProxlessName, dummyNamespaceName)
	assert.Error(t, err)

	helper_createNamespace(t, clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, clientSet)

	// no error - deployment in kubernetes
	deploy, err = removeDeploymentLabel(clientSet, dummyProxlessName, dummyNamespaceName)
	assert.NoError(t, err)

	wantLabels := map[string]string{}
	if !utils.CompareMap(deploy.Labels, wantLabels) {
		t.Errorf("removeDeploymentLabel(); labels = %s, wantLabels = %s", deploy.Labels, wantLabels)
	}
}
