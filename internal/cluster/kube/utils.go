package kube

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/cluster"
	"strings"
)

func genDomains(domains, name, namespace string) []string {
	domainsArray := strings.Split(domains, ",")
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s", name, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s.svc.kubeCluster.local", name, namespace))
	return domainsArray
}

func isAnnotationsProxlessCompatible(meta metav1.ObjectMeta) bool {
	return metav1.HasAnnotation(meta, cluster.AnnotationServiceDomainKey) &&
		metav1.HasAnnotation(meta, cluster.AnnotationServiceDeployKey)
}
