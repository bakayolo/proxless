// TODO this is almost a COPY/PASTE of `kube/service.go` - see how we wanna deal with it

package openshift

import (
	"context"
	"errors"
	"fmt"
	"github.com/openshift/client-go/apps/clientset/versioned"
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
	kubeClientset kubernetes.Interface, ocClientSet versioned.Interface,
	svc *corev1.Service, namespaceScoped bool, proxlessSvc, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[clusterutils.AnnotationServiceDeployKey]

		_, err := createProxlessService(kubeClientset, svc.Name, svc.Namespace, proxlessSvc, proxlessNamespace)

		if err != nil {
			logger.Errorf(err, "Error creating proxless service for %s.%s", svc.Name, svc.Namespace)
			// do not return here - we don't wanna break the proxy forwarding
			// it will be relabel after the informer resync
		}

		_, err = labelDeployment(ocClientSet, deployName, svc.Namespace)

		if err != nil {
			logger.Errorf(err, "Error labelling deployment %s.%s", deployName, svc.Namespace)
			// do not return here - we don't wanna break the proxy forwarding
			// it will be relabel after the informer resync
		}

		port := getPortFromServicePorts(svc.Spec.Ports)
		domains :=
			clusterutils.GenDomains(svc.Annotations[clusterutils.AnnotationServiceDomainKey], svc.Name, svc.Namespace, namespaceScoped)

		ttlSeconds := clusterutils.ParseStringToIntPointer(svc.Annotations[clusterutils.AnnotationServiceTTLSeconds])
		readinessTimeoutSeconds := clusterutils.ParseStringToIntPointer(svc.Annotations[clusterutils.AnnotationServiceReadinessTimeoutSeconds])

		err = upsertMemory(string(svc.UID), svc.Name, port, deployName, svc.Namespace, domains, ttlSeconds, readinessTimeoutSeconds)

		if err == nil {
			logger.Debugf("Service %s.%s added into memory", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error adding service %s.%s into memory", svc.Name, svc.Namespace)
		}
	}
}

func removeServiceFromMemory(
	kubeClientset kubernetes.Interface, ocClientSet versioned.Interface,
	svc *corev1.Service,
	deleteRouteFromMemory func(id string) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[clusterutils.AnnotationServiceDeployKey]

		// we don't process the error here - the deployment might have been delete with the service
		_, _ = removeDeploymentLabel(ocClientSet, deployName, svc.Namespace)

		_ = deleteProxlessService(kubeClientset, svc.Name, svc.Namespace)

		err := deleteRouteFromMemory(string(svc.UID))

		if err == nil {
			logger.Debugf("Service %s.%s removed from memory", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error removing service %s.%s memory", svc.Name, svc.Namespace)
		}
	}
}

func updateServiceMemory(
	kubeClientset kubernetes.Interface, ocClientSet versioned.Interface,
	oldSvc, newSvc *corev1.Service, namespaceScoped bool, proxlessService, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	if clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // updating service
		oldDeployName := oldSvc.Annotations[clusterutils.AnnotationServiceDeployKey]

		if oldDeployName != newSvc.Annotations[clusterutils.AnnotationServiceDeployKey] {
			_, err := removeDeploymentLabel(ocClientSet, oldDeployName, oldSvc.Namespace)
			if err != nil {
				logger.Errorf(err, "error remove proxless label from deployment %s.%s",
					oldDeployName, oldSvc.Namespace)
			}
		}

		// the `addServiceToMemory` is idempotent so we can reuse it in the update
		addServiceToMemory(kubeClientset, ocClientSet, newSvc, namespaceScoped, proxlessService, proxlessNamespace, upsertMemory)
	} else if !clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // adding new service
		addServiceToMemory(kubeClientset, ocClientSet, newSvc, namespaceScoped, proxlessService, proxlessNamespace, upsertMemory)
	} else if clusterutils.IsAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		!clusterutils.IsAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // removing service
		removeServiceFromMemory(kubeClientset, ocClientSet, oldSvc, deleteRouteFromMemory)
	}
}
