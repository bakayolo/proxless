package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/utils"
	"testing"
)

func Test_GenDomains(t *testing.T) {
	testCases := []struct {
		domains, svcName, namespace string
		namespaceScoped             bool
		want                        []string
	}{
		{
			domains:         "example.io",
			svcName:         "dummySvcName",
			namespace:       "dummyNsName",
			namespaceScoped: false,
			want: []string{
				"example.io",
				"dummySvcName.dummyNsName",
				"dummySvcName-proxless.dummyNsName",
				"dummySvcName.dummyNsName.svc.cluster.local",
				"dummySvcName-proxless.dummyNsName.svc.cluster.local",
			},
		},
		{
			domains:         "example.io",
			svcName:         "dummySvcName",
			namespace:       "dummyNsName",
			namespaceScoped: true,
			want: []string{
				"example.io",
				"dummySvcName",
				"dummySvcName-proxless",
				"dummySvcName.dummyNsName",
				"dummySvcName-proxless.dummyNsName",
				"dummySvcName.dummyNsName.svc.cluster.local",
				"dummySvcName-proxless.dummyNsName.svc.cluster.local",
			},
		},
		{

			domains:         "example.io,example.com",
			svcName:         "dummySvcName",
			namespace:       "dummyNsName",
			namespaceScoped: false,
			want: []string{
				"example.io",
				"example.com",
				"dummySvcName.dummyNsName",
				"dummySvcName-proxless.dummyNsName",
				"dummySvcName.dummyNsName.svc.cluster.local",
				"dummySvcName-proxless.dummyNsName.svc.cluster.local",
			},
		},
	}

	for _, tc := range testCases {
		got := GenDomains(tc.domains, tc.svcName, tc.namespace, tc.namespaceScoped)

		if !utils.CompareUnorderedArray(tc.want, got) {
			t.Errorf("genDomains(%s, %s, %s, %t) = %s; want = %s",
				tc.domains, tc.svcName, tc.namespace, tc.namespaceScoped, got, tc.want)
		}
	}
}

func Test_IsAnnotationsProxlessCompatible(t *testing.T) {
	testCases := []struct {
		annotations map[string]string
		want        bool
	}{
		{
			map[string]string{
				AnnotationServiceDomainKey: "domain",
				AnnotationServiceDeployKey: "deploy",
			},
			true,
		},
		{
			map[string]string{
				AnnotationServiceDomainKey: "domain",
			},
			false,
		},
		{
			map[string]string{
				AnnotationServiceDeployKey: "deploy",
			},
			true,
		},
	}

	for _, tc := range testCases {
		got := IsAnnotationsProxlessCompatible(metav1.ObjectMeta{Annotations: tc.annotations})

		if got != tc.want {
			t.Errorf("isAnnotationsProxlessCompatible(%v) = %t; want = %t",
				tc.annotations, got, tc.want)
		}
	}
}
