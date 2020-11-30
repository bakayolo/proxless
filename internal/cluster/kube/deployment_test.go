package kube

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

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
