package kube

import (
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kube-proxless/internal/cluster"
	"testing"
)

var (
	objectMeta = metav1.ObjectMeta{
		Annotations: map[string]string{
			cluster.AnnotationServiceDomainKey: "domain",
			cluster.AnnotationServiceDeployKey: "deploy",
		},
	}
	ports = []corev1.ServicePort{{TargetPort: intstr.IntOrString{IntVal: 8080}}}
)

func helperLabelDeploymentOK(_, _ string) error {
	return nil
}

func helperLabelDeploymentError(_, _ string) error {
	return errors.New("error while labelling")
}

func helperUpsertStoreOk(_, _, _, _, _ string, _ []string) error {
	return nil
}

func helperUpsertStoreError(_, _, _, _, _ string, _ []string) error {
	return errors.New("error while upserting store")
}

func helperUnlabelDeploymentOK(_, _ string) error {
	return nil
}

func helperUnlabelDeploymentError(_, _ string) error {
	return errors.New("error while unlabelling")
}

func helperDeleteRouteFromStoreOK(_ string) error {
	return nil
}

func helperDeleteRouteFromStoreError(_ string) error {
	return errors.New("error while deleting route")
}

func Test_addServiceToStore(t *testing.T) {
	testCases := []struct {
		svc             *corev1.Service
		labelDeployment func(deployName, namespace string) error
		upsertStore     func(id, name, port, deployName, namespace string, domains []string) error
	}{
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUpsertStoreOk,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentError,
			helperUpsertStoreOk,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUpsertStoreError,
		},
	}

	// making sure it never panic
	for _, tc := range testCases {
		addServiceToStore(tc.svc, tc.labelDeployment, tc.upsertStore)
	}
}

func Test_updateServiceInStore(t *testing.T) {
	testCases := []struct {
		oldSvc, newSvc       *corev1.Service
		labelDeployment      func(deployName, namespace string) error
		unlabelDeployment    func(deployName, namespace string) error
		upsertStore          func(id, name, port, deployName, namespace string, domains []string) error
		deleteRouteFromStore func(id string) error
	}{
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUnlabelDeploymentOK,
			helperUpsertStoreOk,
			helperDeleteRouteFromStoreOK,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentError,
			helperUnlabelDeploymentOK,
			helperUpsertStoreOk,
			helperDeleteRouteFromStoreOK,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUnlabelDeploymentError,
			helperUpsertStoreOk,
			helperDeleteRouteFromStoreOK,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUnlabelDeploymentOK,
			helperUpsertStoreError,
			helperDeleteRouteFromStoreOK,
		},
		{ // delete service
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{Ports: ports},
			},
			helperLabelDeploymentOK,
			helperUnlabelDeploymentOK,
			helperUpsertStoreOk,
			helperDeleteRouteFromStoreOK,
		},
		{ // add service
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{Ports: ports},
			},
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperLabelDeploymentOK,
			helperUnlabelDeploymentOK,
			helperUpsertStoreOk,
			helperDeleteRouteFromStoreOK,
		},
	}

	// making sure it never panic
	for _, tc := range testCases {
		updateServiceInStore(
			tc.oldSvc, tc.newSvc,
			tc.labelDeployment, tc.unlabelDeployment,
			tc.upsertStore, tc.deleteRouteFromStore,
		)
	}
}

func Test_removeServiceFromStore(t *testing.T) {
	testCases := []struct {
		svc                  *corev1.Service
		unlabelDeployment    func(deployName, namespace string) error
		deleteRouteFromStore func(id string) error
	}{
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperUnlabelDeploymentOK,
			helperDeleteRouteFromStoreOK,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperUnlabelDeploymentError,
			helperDeleteRouteFromStoreOK,
		},
		{
			&corev1.Service{ObjectMeta: objectMeta, Spec: corev1.ServiceSpec{Ports: ports}},
			helperUnlabelDeploymentOK,
			helperDeleteRouteFromStoreError,
		},
	}

	// making sure it never panic
	for _, tc := range testCases {
		removeServiceFromStore(tc.svc, tc.unlabelDeployment, tc.deleteRouteFromStore)
	}
}
