package kube

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
	clusterutils "kube-proxless/internal/cluster/utils"
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

func TestClusterClient_ScaleUpDeployment(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	client := NewCluster(clientSet, 2)

	timeout := 1

	// error - deployment is not in kubernetes
	assert.Error(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))

	helper_createNamespace(t, clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, clientSet)

	// error - deployment in kubernetes but not available
	assert.Error(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))

	deploy.Status.AvailableReplicas = 1
	helper_updateDeployment(t, clientSet, deploy)

	// no error - deployment in kubernetes and available
	assert.NoError(t, client.ScaleUpDeployment(dummyProxlessName, dummyNamespaceName, timeout))
}

func TestClusterClient_ScaleDownDeployments(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	client := NewCluster(clientSet, 2)

	helper_createNamespace(t, clientSet)
	deploy := helper_createProxlessCompatibleDeployment(t, clientSet)
	randomDeployCreated := helper_createRandomDeployment(t, clientSet) // this deployment must not be scaled down

	// no error - deployment in kubernetes and scaled down
	helper_assertNoError(t, client.ScaleDownDeployments(dummyNamespaceName, helper_shouldScaleDown))

	deploy.Spec.Replicas = pointer.Int32Ptr(1)
	helper_updateDeployment(t, clientSet, deploy)

	// no error - deployment in kubernetes and scaled up
	helper_assertNoError(t, client.ScaleDownDeployments(dummyNamespaceName, helper_shouldScaleDown))

	randomDeploy, _ := getDeployment(clientSet, dummyNonProxlessName, dummyNamespaceName)
	if *randomDeploy.Spec.Replicas != *randomDeployCreated.Spec.Replicas {
		t.Errorf("ScaleDownDeployments(); must not scale down not proxless deployment. Replicas = %d; Want = %d",
			*randomDeploy.Spec.Replicas, *randomDeployCreated.Spec.Replicas)
	}
}

// TODO split this test function - too much sh** here
// the `time.sleep` are here to wait for the informer to sync
func TestClusterClient_RunServicesEngine(t *testing.T) {
	clientSet := fake.NewSimpleClientset()
	servicesInformerResyncInterval := 2
	client := NewCluster(clientSet, servicesInformerResyncInterval)

	memory := fakeMemory{m: map[string]string{}}

	helper_createNamespace(t, clientSet)
	helper_createProxlessCompatibleDeployment(t, clientSet)

	// TODO check how we wanna deal with closing the channel and stopping the routine
	// We could use a context https://github.com/kubernetes/client-go/blob/master/examples/fake-client/main_test.go
	// but not sure if it is worth it
	go client.RunServicesEngine(
		dummyNamespaceName, dummyProxlessName, dummyProxlessName,
		memory.helper_upsertMemory, memory.helper_deleteRouteFromMemory)

	// don't add random services in memory
	helper_createRandomService(t, clientSet)
	time.Sleep(1 * time.Second)
	if len(memory.m) > 0 {
		t.Errorf("RunServicesEngine(); must not add random service information into memory")
	}

	// add proxless compatible services into memory
	service := helper_createProxlessCompatibleService(t, clientSet)
	time.Sleep(1 * time.Second)
	if _, ok := memory.m[string(service.UID)]; !ok {
		t.Errorf("RunServicesEngine(); service not added in memory")
	}
	_, err :=
		clientSet.CoreV1().Services(dummyNamespaceName).Get(
			context.TODO(), clusterutils.GenServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.NoError(t, err)

	// the deployment was not here during creation of the service so proxless label has not been added
	// however the services informer resync must label it
	helper_createRandomDeployment(t, clientSet)
	time.Sleep(time.Duration(servicesInformerResyncInterval) * time.Second)
	randomDeploy, _ := getDeployment(clientSet, dummyNonProxlessName, dummyNamespaceName)
	labelsWant := map[string]string{"app": dummyNonProxlessName, clusterutils.LabelDeploymentProxless: "true"}
	if !utils.CompareMap(randomDeploy.Labels, labelsWant) {
		t.Errorf("RunServicesEngine(); deployment must have the label; labels = %s; labelsWant = %s",
			randomDeploy.Labels, labelsWant)
	}

	// must remove the label from the other deployment
	service.Annotations[clusterutils.AnnotationServiceDeployKey] = dummyProxlessName
	helper_updateService(t, clientSet, service)
	time.Sleep(1 * time.Second)
	randomDeploy, _ = getDeployment(clientSet, dummyNonProxlessName, dummyNamespaceName)
	if _, ok := randomDeploy.Labels[clusterutils.LabelDeploymentProxless]; ok {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", randomDeploy.Labels)
	}

	// must remove the service from the memory and remove the label from the deployment
	// if the service is not proxless compatible anymore
	service.Annotations = map[string]string{}
	helper_updateService(t, clientSet, service)
	time.Sleep(1 * time.Second)
	proxlessDeploy, _ := getDeployment(clientSet, dummyProxlessName, dummyNamespaceName)
	if len(proxlessDeploy.Labels) > 0 {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", proxlessDeploy.Labels)
	}
	if len(memory.m) > 0 {
		t.Errorf("RunServicesEngine(); the service must be removed from the memory")
	}
	_, err =
		clientSet.CoreV1().Services(dummyNamespaceName).Get(
			context.TODO(), clusterutils.GenServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.Error(t, err)

	// must remove the service from the memory and remove the label from the deployment
	// if the service is deleted from kubernetes
	service.Annotations = map[string]string{
		clusterutils.AnnotationServiceDomainKey: "dummy.io",
		clusterutils.AnnotationServiceDeployKey: dummyNonProxlessName,
	}
	helper_updateService(t, clientSet, service)
	_ = clientSet.CoreV1().Services(dummyNamespaceName).Delete(
		context.TODO(), dummyProxlessName, v1.DeleteOptions{})
	time.Sleep(1 * time.Second)
	proxlessDeploy, _ = getDeployment(clientSet, dummyProxlessName, dummyNamespaceName)
	if len(proxlessDeploy.Labels) > 0 {
		t.Errorf("RunServicesEngine(); labels must be removed; labels = %s", proxlessDeploy.Labels)
	}
	if len(memory.m) > 0 {
		t.Errorf("RunServicesEngine(); the service must be removed from the memory")
	}
	_, err =
		clientSet.CoreV1().Services(dummyNamespaceName).Get(
			context.TODO(), clusterutils.GenServiceToAppName(dummyProxlessName), v1.GetOptions{})
	assert.Error(t, err)
}
