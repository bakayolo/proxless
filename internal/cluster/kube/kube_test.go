package kube

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

func TestClusterClient_ScaleUpDeployment(t *testing.T) {
	client := NewCluster(fake.NewSimpleClientset())

	timeout := 1

	// error - deployment is not in kubernetes
	assert.Error(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))

	helper_createNamespace(t, client.clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, client.clientSet)

	// error - deployment in kubernetes but not available
	assert.Error(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))

	deploy.Status.AvailableReplicas = 1
	helper_updateDeployment(t, client.clientSet, deploy)

	// no error - deployment in kubernetes and available
	assert.NoError(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))
}

func TestClusterClient_ScaleDownDeployments(t *testing.T) {
	client := NewCluster(fake.NewSimpleClientset())

	// error - deployment is not in kubernetes
	helper_assertAtLeastOneError(t, client.ScaleDownDeployments(dummyNamespaceName, helper_shouldScaleDown))

	helper_createNamespace(t, client.clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, client.clientSet)
	randomDeployCreated := helper_createRandomDeployment(t, client.clientSet) // this deployment must not be scaled down

	// no error - deployment in kubernetes and scaled down
	helper_assertNoError(t, client.ScaleDownDeployments(dummyNamespaceName, helper_shouldScaleDown))

	deploy.Spec.Replicas = pointer.Int32Ptr(1)
	helper_updateDeployment(t, client.clientSet, deploy)

	// no error - deployment in kubernetes and scaled up
	helper_assertNoError(t, client.ScaleDownDeployments(dummyNamespaceName, helper_shouldScaleDown))

	randomDeploy, _ := getDeployment(client.clientSet, dummyNonProxlessName, dummyNamespaceName)
	if *randomDeploy.Spec.Replicas != *randomDeployCreated.Spec.Replicas {
		t.Errorf("ScaleDownDeployments(); must not scale down not proxless deployment. Replicas = %d; Want = %d",
			*randomDeploy.Spec.Replicas, *randomDeployCreated.Spec.Replicas)
	}
}

// TODO split this test function - too much sh** here
// the `time.sleep` are here to wait for the informer to sync
func TestClusterClient_RunServicesEngine(t *testing.T) {
	client := NewCluster(fake.NewSimpleClientset())
	client.servicesInformerResyncInterval = 2

	store := fakeStore{store: map[string]string{}}

	helper_createNamespace(t, client.clientSet)
	helper_createProxlessCompatibleDeployment(t, client.clientSet)

	// TODO check how we wanna deal with closing the channel and stopping the routine
	// We could use a context https://github.com/kubernetes/client-go/blob/master/examples/fake-client/main_test.go
	// but not sure if it is worth it
	go client.RunServicesEngine(
		dummyNamespaceName, dummyProxlessName, dummyProxlessName,
		store.helper_upsertStore, store.helper_deleteRouteFromStore)

	// don't store random services
	helper_createRandomService(t, client.clientSet)
	time.Sleep(1 * time.Second)
	if len(store.store) > 0 {
		t.Errorf("RunServicesEngine(); must not store random service information")
	}

	// store proxless compatible services
	service := helper_createProxlessCompatibleService(t, client.clientSet)
	time.Sleep(1 * time.Second)
	if _, ok := store.store[string(service.UID)]; !ok {
		t.Errorf("RunServicesEngine(); service not added in store")
	}
	_, err :=
		client.clientSet.CoreV1().Services(dummyNamespaceName).Get(genServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.NoError(t, err)

	// the deployment was not here during creation of the service so proxless label has not been added
	// however the services informer resync must label it
	helper_createRandomDeployment(t, client.clientSet)
	time.Sleep(time.Duration(client.servicesInformerResyncInterval) * time.Second)
	randomDeploy, _ := getDeployment(client.clientSet, dummyNonProxlessName, dummyNamespaceName)
	labelsWant := map[string]string{cluster.LabelDeploymentProxless: "true"}
	if !utils.CompareMap(randomDeploy.Labels, labelsWant) {
		t.Errorf("RunServicesEngine(); deployment must have the label; labels = %s; labelsWant = %s",
			randomDeploy.Labels, labelsWant)
	}

	// must remove the label from the other deployment
	service.Annotations[cluster.AnnotationServiceDeployKey] = dummyProxlessName
	helper_updateService(t, client.clientSet, service)
	time.Sleep(1 * time.Second)
	randomDeploy, _ = getDeployment(client.clientSet, dummyNonProxlessName, dummyNamespaceName)
	if len(randomDeploy.Labels) > 0 {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", randomDeploy.Labels)
	}

	// must remove the service from the store and remove the label from the deployment
	// if the service is not proxless compatible anymore
	service.Annotations = map[string]string{}
	helper_updateService(t, client.clientSet, service)
	time.Sleep(1 * time.Second)
	proxlessDeploy, _ := getDeployment(client.clientSet, dummyProxlessName, dummyNamespaceName)
	if len(proxlessDeploy.Labels) > 0 {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", proxlessDeploy.Labels)
	}
	if len(store.store) > 0 {
		t.Errorf("RunServicesEngine(); the service must be removed from the store")
	}
	_, err =
		client.clientSet.CoreV1().Services(dummyNamespaceName).Get(genServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.Error(t, err)

	// must remove the service from the store and remove the label from the deployment
	// if the service is deleted from kubernetes
	service.Annotations = map[string]string{
		cluster.AnnotationServiceDomainKey: "dummy.io",
		cluster.AnnotationServiceDeployKey: dummyNonProxlessName,
	}
	helper_updateService(t, client.clientSet, service)
	_ = client.clientSet.CoreV1().Services(dummyNamespaceName).Delete(dummyProxlessName, &v1.DeleteOptions{})
	time.Sleep(1 * time.Second)
	proxlessDeploy, _ = getDeployment(client.clientSet, dummyProxlessName, dummyNamespaceName)
	if len(proxlessDeploy.Labels) > 0 {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", proxlessDeploy.Labels)
	}
	if len(store.store) > 0 {
		t.Errorf("RunServicesEngine(); the service must be removed from the store")
	}
	_, err =
		client.clientSet.CoreV1().Services(dummyNamespaceName).Get(genServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.Error(t, err)
}
