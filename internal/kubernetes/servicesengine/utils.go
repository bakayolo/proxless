package servicesengine

import (
	"errors"
	"fmt"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"kube-proxless/internal/kubernetes"
	"strconv"
	"strings"
	"time"
)

func stringifyUid(uid types.UID) string {
	return string(uid)
}

func parseService(obj interface{}) (*core.Service, error) {
	svc, ok := obj.(*core.Service)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Event for invalid object; got %T want *core.Service", obj))
	}
	return svc, nil
}

func genInformerFactory(namespace string) informers.SharedInformerFactory {
	opts := make([]informers.SharedInformerOption, 0)
	if namespace != "" {
		opts = append(opts, informers.WithNamespace(namespace))
	}
	// TODO make the default resync configurable
	return informers.NewSharedInformerFactoryWithOptions(kubernetes.ClientSet, 30*time.Second, opts...)
}

func genDomains(domains, name, namespace string) []string {
	domainsArray := strings.Split(domains, ",")
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s", name, namespace))
	domainsArray = append(domainsArray, fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace))
	return domainsArray
}

func genPort(ports []core.ServicePort) string {
	port := ports[0] // TODO add possibility to manage multiple ports
	return strconv.Itoa(int(port.TargetPort.IntVal))
}

func isProxlessCompatible(svc core.Service) bool {
	return metav1.HasAnnotation(svc.ObjectMeta, annotationDomainKey) &&
		metav1.HasAnnotation(svc.ObjectMeta, annotationDeployKey) &&
		len(svc.Spec.Ports) == 1 // TODO add possibility to manage multiple ports
}
