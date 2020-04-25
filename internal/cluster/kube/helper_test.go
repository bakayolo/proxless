package kube

import (
	"errors"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"kube-proxless/internal/cluster"
	"testing"
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
	_, err := clientSet.CoreV1().Namespaces().Create(dummyNamespace)
	assert.NoError(t, err)
}

func helper_createProxlessCompatibleDeployment(t *testing.T, clientSet kubernetes.Interface) *appsv1.Deployment {
	dummyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyProxlessName,
			Namespace: dummyNamespaceName,
			Labels:    map[string]string{cluster.LabelDeploymentProxless: "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(0),
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 0,
		},
	}
	deploy, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Create(dummyDeploy)
	assert.NoError(t, err)
	return deploy
}

func helper_createRandomDeployment(t *testing.T, clientSet kubernetes.Interface) *appsv1.Deployment {
	dummyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyNonProxlessName,
			Namespace: dummyNamespaceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}
	deploy, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Create(dummyDeploy)
	assert.NoError(t, err)
	return deploy
}

func helper_updateDeployment(
	t *testing.T, clientSet kubernetes.Interface, deploy *appsv1.Deployment) *appsv1.Deployment {
	deployUpdated, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Update(deploy)
	assert.NoError(t, err)
	return deployUpdated
}

func helper_createProxlessCompatibleService(t *testing.T, clientSet kubernetes.Interface) *corev1.Service {
	dummyService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyProxlessName,
			Namespace: dummyNamespaceName,
			Annotations: map[string]string{
				cluster.AnnotationServiceDomainKey: "dummy.io",
				// the deployment will be the one without the proxless labels since it has to be labelled by
				// the services informer
				cluster.AnnotationServiceDeployKey: dummyNonProxlessName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{TargetPort: intstr.IntOrString{IntVal: 8080}},
			},
		},
	}
	deploy, err := clientSet.CoreV1().Services(dummyNamespaceName).Create(dummyService)
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
	deploy, err := clientSet.CoreV1().Services(dummyNamespaceName).Create(dummyService)
	assert.NoError(t, err)
	return deploy
}

func helper_updateService(t *testing.T, clientSet kubernetes.Interface, svc *corev1.Service) *corev1.Service {
	svcUpdated, err := clientSet.CoreV1().Services(dummyNamespaceName).Update(svc)
	assert.NoError(t, err)
	return svcUpdated
}

func helper_assertAtLeastOneError(t *testing.T, errs []error) {
	if errs == nil || len(errs) == 0 {
		t.Errorf("Array must have at least an error")
	}
}

func helper_assertNoError(t *testing.T, errs []error) {
	if errs != nil && len(errs) > 0 {
		t.Errorf("Array must not have any error")
	}
}

func helper_shouldScaleDown(_, _ string) (bool, error) {
	return true, nil
}

type fakeStore struct {
	store map[string]string
}

func (s *fakeStore) helper_upsertStore(id, name, port, deployName, namespace string, domains []string) error {
	if deployName == "" {
		return errors.New("error upserting store")
	}
	s.store[id] = deployName
	return nil
}

func (s *fakeStore) helper_deleteRouteFromStore(id string) error {
	if _, ok := s.store[id]; ok {
		delete(s.store, id)
		return nil
	}
	return errors.New("route not found")
}
