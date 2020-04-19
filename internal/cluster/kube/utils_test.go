package kube

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/utils"
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

func Test_genDomains(t *testing.T) {
	testCases := []struct {
		domains, svcName, namespace string
		want                        []string
	}{
		{"example.io", "svc", "ns",
			[]string{"example.io", "svc.ns", "svc.ns.svc.cluster.local"},
		},
		{"example.io,example.com", "svc", "ns",
			[]string{"example.io", "example.com", "svc.ns", "svc.ns.svc.cluster.local"},
		},
	}

	for _, tc := range testCases {
		got := genDomains(tc.domains, tc.svcName, tc.namespace)

		if !utils.CompareUnorderedArray(tc.want, got) {
			t.Errorf("genDomains(%s, %s, %s) = %s; want = %s",
				tc.domains, tc.svcName, tc.namespace, got, tc.want)
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

func Test_isAnnotationsProxlessCompatible(t *testing.T) {
	testCases := []struct {
		annotations map[string]string
		want        bool
	}{
		{
			map[string]string{
				cluster.AnnotationServiceDomainKey: "domain",
				cluster.AnnotationServiceDeployKey: "deploy",
			},
			true,
		},
		{
			map[string]string{
				cluster.AnnotationServiceDomainKey: "domain",
			},
			false,
		},
		{
			map[string]string{
				cluster.AnnotationServiceDeployKey: "deploy",
			},
			false,
		},
	}

	for _, tc := range testCases {
		got := isAnnotationsProxlessCompatible(metav1.ObjectMeta{Annotations: tc.annotations})

		if got != tc.want {
			t.Errorf("isAnnotationsProxlessCompatible(%v) = %t; want = %t",
				tc.annotations, got, tc.want)
		}
	}
}
