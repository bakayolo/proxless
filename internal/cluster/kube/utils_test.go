package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/utils"
	"testing"
)

func Test_genDomains(t *testing.T) {
	testCases := []struct {
		domains, svcName, namespace string
		want                        []string
	}{
		{"example.io", "svc", "ns",
			[]string{"example.io", "svc.ns", "svc.ns.svc.kubeCluster.local"},
		},
		{"example.io,example.com", "svc", "ns",
			[]string{"example.io", "example.com", "svc.ns", "svc.ns.svc.kubeCluster.local"},
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
