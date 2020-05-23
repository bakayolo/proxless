package openshift

import (
	ocfake "github.com/openshift/client-go/apps/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	kubefake "k8s.io/client-go/kubernetes/fake"
	utils2 "kube-proxless/internal/cluster/utils"
	"kube-proxless/internal/utils"
	"testing"
)

func Test_scaleUpDeployment(t *testing.T) {
	// TestClusterClient_ScaleUpDeployment(t)
}

func Test_waitForDeploymentAvailable(t *testing.T) {
	ocClientSet := ocfake.NewSimpleClientset()
	kubeClientSet := kubefake.NewSimpleClientset()

	timeout := 1

	// error - deployment is not in kubernetes
	assert.Error(t, waitForDeploymentAvailable(ocClientSet, dummyProxlessName, dummyNamespaceName, timeout))

	helper_createNamespace(t, kubeClientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, ocClientSet)

	// error - deployment in kubernetes but not available
	assert.Error(t, waitForDeploymentAvailable(ocClientSet, dummyProxlessName, dummyNamespaceName, timeout))

	deploy.Status.AvailableReplicas = 1
	helper_updateDeployment(t, ocClientSet, deploy)

	// no error - deployment in kubernetes and available
	assert.NoError(t, waitForDeploymentAvailable(ocClientSet, dummyProxlessName, dummyNamespaceName, timeout))
}

func Test_scaleDownDeployments(t *testing.T) {
	// TestClusterClient_ScaleDownDeployments(t)
}

func Test_labelDeployment(t *testing.T) {
	ocClientSet := ocfake.NewSimpleClientset()
	kubeClientSet := kubefake.NewSimpleClientset()

	// error - deployment is not in kubernetes
	_, err := labelDeployment(ocClientSet, dummyNonProxlessName, dummyNamespaceName)
	assert.Error(t, err)

	helper_createNamespace(t, kubeClientSet)
	deploy := helper_createRandomDeployment(t, ocClientSet)

	// no error - deployment in kubernetes
	deploy, err = labelDeployment(ocClientSet, dummyNonProxlessName, dummyNamespaceName)
	assert.NoError(t, err)

	wantLabels := map[string]string{utils2.LabelDeploymentProxless: "true"}
	if !utils.CompareMap(deploy.Labels, wantLabels) {
		t.Errorf("labelDeployment(); labels = %s, wantLabels = %s", deploy.Labels, wantLabels)
	}
}

func Test_removeDeploymentLabel(t *testing.T) {
	ocClientSet := ocfake.NewSimpleClientset()
	kubeClientSet := kubefake.NewSimpleClientset()

	// error - deployment is not in kubernetes
	_, err := removeDeploymentLabel(ocClientSet, dummyProxlessName, dummyNamespaceName)
	assert.Error(t, err)

	helper_createNamespace(t, kubeClientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, ocClientSet)

	// no error - deployment in kubernetes
	deploy, err = removeDeploymentLabel(ocClientSet, dummyProxlessName, dummyNamespaceName)
	assert.NoError(t, err)

	wantLabels := map[string]string{}
	if !utils.CompareMap(deploy.Labels, wantLabels) {
		t.Errorf("removeDeploymentLabel(); labels = %s, wantLabels = %s", deploy.Labels, wantLabels)
	}
}
