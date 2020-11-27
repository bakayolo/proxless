package kube

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clusterutils "kube-proxless/internal/cluster/utils"
	"kube-proxless/internal/logger"
	"strconv"
)

func createProxlessService(
	clientSet kubernetes.Interface, appSvc, appNs, proxlessSvc, proxlessNs string) (*corev1.Service, error) {
	svc, err := clientSet.CoreV1().Services(appNs).Create(context.TODO(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        clusterutils.GenServiceToAppName(appSvc),
			Annotations: map[string]string{"owner": "proxless"},
		},
		Spec: corev1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: fmt.Sprintf("%s.%s.svc.cluster.local", proxlessSvc, proxlessNs),
		},
	}, metav1.CreateOptions{})

	// not an issue if proxless service already exists and we don't wanna update it
	if k8serrors.IsAlreadyExists(err) {
		return svc, nil
	}

	return svc, err
}

func deleteProxlessService(clientSet kubernetes.Interface, appSvc, appNs string) error {
	return clientSet.CoreV1().Services(appNs).Delete(
		context.TODO(), clusterutils.GenServiceToAppName(appSvc), metav1.DeleteOptions{})
}

func parseService(obj interface{}) (*corev1.Service, error) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return nil, errors.New(fmt.Sprintf("event for invalid object; got %T want *core.Service", obj))
	}
	return svc, nil
}

func getPortFromServicePorts(ports []corev1.ServicePort) string {
	port := ports[0] // TODO add possibility to manage multiple ports

	return strconv.Itoa(int(port.TargetPort.IntVal))
}

func addServiceToMemory(
	clientset kubernetes.Interface, svc *corev1.Service, namespaceScoped bool,
	proxlessSvc, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[clusterutils.AnnotationServiceDeployKey]
		domains :=
			clusterutils.GenDomains(svc.Annotations[clusterutils.AnnotationServiceDomainKey], svc.Name, svc.Namespace, namespaceScoped)
		ttlSeconds := clusterutils.ParseStringToIntPointer(svc.Annotations[clusterutils.AnnotationServiceTTLSeconds])
		readinessTimeoutSeconds := clusterutils.ParseStringToIntPointer(svc.Annotations[clusterutils.AnnotationServiceReadinessTimeoutSeconds])

		var err error
		if serviceName, ok := svc.Annotations[clusterutils.AnnotationServiceServiceName]; ok {
			appNs := svc.Namespace
			svc, err = clientset.CoreV1().Services(appNs).Get(context.TODO(), serviceName, metav1.GetOptions{})

			if err != nil {
				logger.Errorf(err, "Error finding service %s.%s", serviceName, appNs)
				return
			}
		} else {
			_, err := createProxlessService(clientset, svc.Name, svc.Namespace, proxlessSvc, proxlessNamespace)

			if err != nil {
				logger.Errorf(err, "Error creating proxless service for %s.%s", svc.Name, svc.Namespace)
				// do not return here - we don't wanna break the proxy forwarding
			}
		}

		port := getPortFromServicePorts(svc.Spec.Ports)

		id := clusterutils.GenRouteId(svc.Name, svc.Namespace)
		err = upsertMemory(id, svc.Name, port, deployName, svc.Namespace, domains, ttlSeconds, readinessTimeoutSeconds)

		if err == nil {
			logger.Debugf("Service %s.%s added into memory", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error adding service %s.%s into memory", svc.Name, svc.Namespace)
		}
	}
}

func removeServiceFromMemory(
	clientset kubernetes.Interface, svc *corev1.Service,
	deleteRouteFromMemory func(id string) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(svc.ObjectMeta) {
		if serviceName, ok := svc.Annotations[clusterutils.AnnotationServiceServiceName]; ok {
			appNs := svc.Namespace
			var err error
			svc, err = clientset.CoreV1().Services(appNs).Get(context.TODO(), serviceName, metav1.GetOptions{})

			if err != nil {
				logger.Errorf(err, "Error finding service %s.%s", serviceName, appNs)
				return
			}
		}

		_ = deleteProxlessService(clientset, svc.Name, svc.Namespace)

		id := clusterutils.GenRouteId(svc.Name, svc.Namespace)
		err := deleteRouteFromMemory(id)

		if err == nil {
			logger.Debugf("Service %s.%s removed from memory", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error removing service %s.%s memory", svc.Name, svc.Namespace)
		}
	}
}

func updateServiceMemory(
	clientset kubernetes.Interface, oldSvc, newSvc *corev1.Service, namespaceScoped bool,
	proxlessService, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // updating service
		// the `addServiceToMemory` is idempotent so we can reuse it in the update
		addServiceToMemory(clientset, newSvc, namespaceScoped, proxlessService, proxlessNamespace, upsertMemory)
	} else if !clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // adding new service
		addServiceToMemory(clientset, newSvc, namespaceScoped, proxlessService, proxlessNamespace, upsertMemory)
	} else if clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		!clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // removing service
		removeServiceFromMemory(clientset, oldSvc, deleteRouteFromMemory)
	}
}
