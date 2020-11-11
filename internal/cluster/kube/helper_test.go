package kube

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
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

func helper_createProxlessCompatibleDeployment(t *testing.T, clientSet kubernetes.Interface) *appsv1.Deployment {
	dummyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyProxlessName,
			Namespace: dummyNamespaceName,
			Labels:    map[string]string{utils.LabelDeploymentProxless: "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(0),
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 0,
		},
	}
	deploy, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Create(
		context.TODO(), dummyDeploy, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_createRandomDeployment(t *testing.T, clientSet kubernetes.Interface) *appsv1.Deployment {
	dummyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummyNonProxlessName,
			Namespace: dummyNamespaceName,
			Labels:    map[string]string{"app": dummyNonProxlessName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}
	deploy, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Create(
		context.TODO(), dummyDeploy, metav1.CreateOptions{})
	assert.NoError(t, err)
	return deploy
}

func helper_updateDeployment(
	t *testing.T, clientSet kubernetes.Interface, deploy *appsv1.Deployment) *appsv1.Deployment {
	deployUpdated, err := clientSet.AppsV1().Deployments(dummyNamespaceName).Update(
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

func (s *fakeMemory) helper_upsertMemory(
	id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error {
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
