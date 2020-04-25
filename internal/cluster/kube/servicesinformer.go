package kube

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/logger"
	"strconv"
	"time"
)

func runServicesInformer(
	clientSet kubernetes.Interface,
	namespace string,
	informerResyncInterval int,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	opts := make([]informers.SharedInformerOption, 0)
	if namespace != "" {
		opts = append(opts, informers.WithNamespace(namespace))
	}
	informer := informers.
		NewSharedInformerFactoryWithOptions(clientSet, time.Duration(informerResyncInterval)*time.Second, opts...).
		Core().V1().Services().Informer()

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc, err := parseService(obj)

			if err != nil {
				logger.Errorf(err, "Cannot process service in AddFunc handler")
				return
			}

			addServiceToStore(clientSet, svc, upsertStore)

			return
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldSvc, err := parseService(oldObj)

			if err != nil {
				logger.Errorf(err, "Cannot process service in UpdateFunc handler")
			}

			newSvc, err := parseService(newObj)

			if err != nil {
				logger.Errorf(err, "Cannot process service in UpdateFunc handler")
			}

			updateServiceInStore(clientSet, oldSvc, newSvc, upsertStore, deleteRouteFromStore)
			return
		},
		DeleteFunc: func(obj interface{}) {
			svc, err := parseService(obj)

			if err != nil {
				logger.Errorf(err, "Cannot process service in DeleteFunc handler")
				return
			}

			removeServiceFromStore(clientSet, svc, deleteRouteFromStore)

			return
		},
	}
	informer.AddEventHandler(eventHandler)

	stopCh := make(chan struct{})
	defer close(stopCh)
	informer.Run(stopCh)
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

func addServiceToStore(
	clientset kubernetes.Interface, svc *corev1.Service,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
) {
	if isAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[cluster.AnnotationServiceDeployKey]

		_, err := labelDeployment(clientset, deployName, svc.Namespace)

		if err != nil {
			logger.Errorf(err, "Error labelling deployment %s.%s", deployName, svc.Namespace)
			// do not return here - we don't wanna break the proxy forwarding
		}

		port := getPortFromServicePorts(svc.Spec.Ports)
		domains := genDomains(svc.Annotations[cluster.AnnotationServiceDomainKey], svc.Name, svc.Namespace)

		err = upsertStore(string(svc.UID), svc.Name, port, deployName, svc.Namespace, domains)

		if err == nil {
			logger.Debugf("Service %s.%s added into the store", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error adding service %s.%s into the store", svc.Name, svc.Namespace)
		}
	}
}

func removeServiceFromStore(
	clientset kubernetes.Interface, svc *corev1.Service,
	deleteRouteFromStore func(id string) error,
) {
	if isAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[cluster.AnnotationServiceDeployKey]

		// we don't process the error here - the deployment might have been delete with the service
		_, _ = unlabelDeployment(clientset, deployName, svc.Namespace)

		err := deleteRouteFromStore(string(svc.UID))

		if err == nil {
			logger.Debugf("Service %s.%s removed from the store", svc.Name, svc.Namespace)
		} else {
			logger.Errorf(err, "Error removing service %s.%s from store", svc.Name, svc.Namespace)
		}
	}
}

func updateServiceInStore(
	clientset kubernetes.Interface, oldSvc, newSvc *corev1.Service,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	if isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // updating service
		oldDeployName := oldSvc.Annotations[cluster.AnnotationServiceDeployKey]

		if oldDeployName != newSvc.Annotations[cluster.AnnotationServiceDeployKey] {
			_, err := unlabelDeployment(clientset, oldDeployName, oldSvc.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf(
					"error unlabelling deployment %s.%s", oldDeployName, oldSvc.Namespace)
			}
		}

		// the `addServiceToStore` is idempotent so we can reuse it in the update
		addServiceToStore(clientset, newSvc, upsertStore)
	} else if !isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // adding new service
		addServiceToStore(clientset, newSvc, upsertStore)
	} else if isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) &&
		!isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // removing service
		removeServiceFromStore(clientset, oldSvc, deleteRouteFromStore)
	}
}
