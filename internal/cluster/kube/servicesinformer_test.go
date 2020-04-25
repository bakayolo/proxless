package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func Test_parseService(t *testing.T) {
	testCases := []struct {
		svc       interface{}
		errWanted bool
	}{
		{&corev1.Service{}, false},
		{&corev1.ConfigMap{}, true},
	}

	for _, tc := range testCases {
		_, errGot := parseService(tc.svc)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("parseService(%v) = %v; error wanted = %t",
				tc.svc, errGot, tc.errWanted)
		}
	}
}

func Test_getPortFromServicePorts(t *testing.T) {
	testCases := []struct {
		port []corev1.ServicePort
		want string
	}{
		{[]corev1.ServicePort{{TargetPort: intstr.IntOrString{IntVal: 80}}}, "80"},
		{[]corev1.ServicePort{{TargetPort: intstr.IntOrString{IntVal: 8080}}, {TargetPort: intstr.IntOrString{IntVal: 80}}}, "8080"},
	}

	for _, tc := range testCases {
		got := getPortFromServicePorts(tc.port)

		if got != tc.want {
			t.Errorf("getPortFromServicePorts(%v) = %s; want = %s",
				tc.port, got, tc.want)
		}
	}
}

func Test_runServicesInformer(t *testing.T) {
	TestClusterClient_RunServicesEngine(t)
}
