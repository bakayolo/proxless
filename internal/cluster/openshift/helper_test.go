// TODO this is almost a COPY/PASTE of `kube/service.go` - see how we wanna deal with it

package openshift

import (
	"context"
	"errors"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/client-go/apps/clientset/versioned"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"kube-proxless/internal/cluster/utils"
	"testing"
	"time"
)

const (
	dummyNamespaceName   = "dummy-namespace"
	dummyNonProxlessName = "dummy-non-proxless"
	dummyProxlessName    = "dummy-proxless"
)

func helper_createNamespace(t *testing.T, clientSet kubernetes.Interface) {
	dummyNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: dummyNamespaceName,
		},
	}
	_, err := clientSet.CoreV1().Namespaces().Create(
		context.TODO(), dummyNamespace, metav1.CreateOptions{})
	assert.NoError(t, err)
}

func helper_createProxlessCompatibleDeployment(t *testing.T, clientSet versioned.Interface) *appsv1.DeploymentConfig {
	dummyDeploy := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyProxlessName,
			Namespace: dummyNamespaceName,
			Labels:    map[string]string{utils.LabelDeploymentProxless: "true"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 0,
		},
		Status: appsv1.DeploymentConfigStatus{
			AvailableReplicas: 0,
		},
	}
	deploy, err := clientSet.AppsV1().DeploymentConfigs(dummyNamespaceName).Create(
		context.TODO(), dummyDeploy, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_createRandomDeployment(t *testing.T, clientSet versioned.Interface) *appsv1.DeploymentConfig {
	dummyDeploy := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyNonProxlessName,
			Namespace: dummyNamespaceName,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
		},
		Status: appsv1.DeploymentConfigStatus{
			AvailableReplicas: 1,
		},
	}
	deploy, err := clientSet.AppsV1().DeploymentConfigs(dummyNamespaceName).Create(
		context.TODO(), dummyDeploy, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_updateDeployment(
	t *testing.T, clientSet versioned.Interface, deploy *appsv1.DeploymentConfig) *appsv1.DeploymentConfig {

	deployUpdated, err := clientSet.AppsV1().DeploymentConfigs(dummyNamespaceName).Update(
		context.TODO(), deploy, metav1.UpdateOptions{})
	assert.NoError(t, err)
	return deployUpdated
}

func helper_createProxlessCompatibleService(t *testing.T, clientSet kubernetes.Interface) *corev1.Service {
	dummyService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyProxlessName,
			Namespace: dummyNamespaceName,
			Annotations: map[string]string{
				utils.AnnotationServiceDomainKey: "dummy.io",
				// the deployment will be the one without the proxless labels since it has to be labelled by
				// the services informer
				utils.AnnotationServiceDeployKey: dummyNonProxlessName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{TargetPort: intstr.IntOrString{IntVal: 8080}},
			},
		},
	}
	deploy, err := clientSet.CoreV1().Services(dummyNamespaceName).Create(
		context.TODO(), dummyService, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_createRandomService(t *testing.T, clientSet kubernetes.Interface) *corev1.Service {
	dummyService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyNonProxlessName,
			Namespace: dummyNamespaceName,
		},
	}
	deploy, err := clientSet.CoreV1().Services(dummyNamespaceName).Create(
		context.TODO(), dummyService, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_updateService(t *testing.T, clientSet kubernetes.Interface, svc *corev1.Service) *corev1.Service {
	svcUpdated, err := clientSet.CoreV1().Services(dummyNamespaceName).Update(
		context.TODO(), svc, metav1.UpdateOptions{})
	assert.NoError(t, err)
	return svcUpdated
}

func helper_assertNoError(t *testing.T, errs []error) {
	if errs != nil && len(errs) > 0 {
		t.Errorf("Array must not have any error")
	}
}

func helper_shouldScaleDown(_, _ string) (bool, time.Duration, error) {
	return true, time.Now().Sub(time.Now()), nil
}

type fakeMemory struct {
	m map[string]string
}

func (s *fakeMemory) helper_upsertMemory(id, name, port, deployName, namespace string, domains []string) error {
	if deployName == "" {
		return errors.New("error upserting m")
	}
	s.m[id] = deployName
	return nil
}

func (s *fakeMemory) helper_deleteRouteFromMemory(id string) error {
	if _, ok := s.m[id]; ok {
		delete(s.m, id)
		return nil
	}
	return errors.New("route not found")
}
