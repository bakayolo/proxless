package kube

import (
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-proxless/internal/cluster"
	"strconv"
	"strings"
)

func parseService(obj interface{}) (*corev1.Service, error) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return nil, errors.New(fmt.Sprintf("event for invalid object; got %T want *core.Service", obj))
	}
	return svc, nil
}

func genDomains(domains, name, namespace string) []string {
	domainsArray := strings.Split(domains, ",")
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s", name, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace))
	return domainsArray
}

func getPortFromServicePorts(ports []corev1.ServicePort) string {
	port := ports[0] // TODO add possibility to manage multiple ports
	return strconv.Itoa(int(port.TargetPort.IntVal))
}

func isAnnotationsProxlessCompatible(meta metav1.ObjectMeta) bool {
	return metav1.HasAnnotation(meta, cluster.AnnotationServiceDomainKey) &&
		metav1.HasAnnotation(meta, cluster.AnnotationServiceDeployKey)
}
