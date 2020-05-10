package kube

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/cluster"
	"strings"
)

func genServiceToAppName(svcName string) string {
	return fmt.Sprintf("%s-proxless", svcName)
}

func genDomains(domains, name, namespace string, namespaceScoped bool) []string {
	svcName := genServiceToAppName(name)
	var domainsArray []string
	if domains != "" {
		domainsArray = strings.Split(domains, ",")
	}
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s", name, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s", svcName, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s.svc.cluster.local", svcName, namespace))

	if namespaceScoped {
		domainsArray = append(domainsArray, name)
		domainsArray = append(domainsArray, svcName)
	}

	return domainsArray
}

func isAnnotationsProxlessCompatible(meta metav1.ObjectMeta) bool {
	return metav1.HasAnnotation(meta, cluster.AnnotationServiceDeployKey)
}
