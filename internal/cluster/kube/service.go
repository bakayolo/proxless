package kube

import (
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"kube-proxless/internal/cluster"
	"time"
)

type KubeServiceInterface interface {
	runServicesInformer(
		namespace string,
		labelDeployment func(deployName, namespace string) error,
		unlabelDeployment func(deployName, namespace string) error,
		upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
		deleteRouteFromStore func(id string) error,
	)
}

type KubeServiceClient struct {
	clientSet *kubernetes.Clientset
}

func (c *KubeServiceClient) genInformerFactory(namespace string) cache.SharedIndexInformer {
	opts := make([]informers.SharedInformerOption, 0)
	if namespace != "" {
		opts = append(opts, informers.WithNamespace(namespace))
	}
	// TODO make the default resync configurable
	informer := informers.
		NewSharedInformerFactoryWithOptions(c.clientSet, 30*time.Second, opts...).
		Core().V1().Services().Informer()
	return informer
}

// TODO I probably went overboard with the visitor pattern.
func (c *KubeServiceClient) runServicesInformer(
	namespace string,
	labelDeployment func(deployName, namespace string) error,
	unlabelDeployment func(deployName, namespace string) error,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	informer := c.genInformerFactory(namespace)
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc, err := parseService(obj)
			if err != nil {
				log.Error().Err(err).Msgf("Cannot process service in AddFunc handler")
				return
			}
			addServiceToStore(svc, labelDeployment, upsertStore)
			return
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldSvc, err0 := parseService(oldObj)
			if err0 != nil {
				log.Error().Err(err0).Msgf("Cannot process service in AddFunc handler")
			}

			newSvc, err1 := parseService(newObj)
			if err1 != nil {
				log.Error().Err(err1).Msgf("Cannot process service in AddFunc handler")
			}

			updateServiceInStore(oldSvc, newSvc, labelDeployment, unlabelDeployment, upsertStore, deleteRouteFromStore)
			return
		},
		DeleteFunc: func(obj interface{}) {
			svc, err := parseService(obj)
			if err != nil {
				log.Error().Err(err).Msgf("Cannot process service in AddFunc handler")
				return
			}
			removeServiceFromStore(svc, unlabelDeployment, deleteRouteFromStore)
			return
		},
	}
	informer.AddEventHandler(eventHandler)

	log.Info().Msgf("Starting Services Informer")

	stopCh := make(chan struct{})
	defer close(stopCh)
	informer.Run(stopCh)
}

func addServiceToStore(
	svc *corev1.Service,
	labelDeployment func(deployName, namespace string) error,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
) {
	if isAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[cluster.AnnotationServiceDeployKey]

		err := labelDeployment(deployName, svc.Namespace)

		if err != nil {
			log.Error().Err(err).Msgf("error labelling deployment %s.%s", deployName, svc.Namespace)
		}

		port := getPortFromServicePorts(svc.Spec.Ports)
		domains := genDomains(svc.Annotations[cluster.AnnotationServiceDomainKey], svc.Name, svc.Namespace)

		err = upsertStore(string(svc.UID), svc.Name, port, deployName, svc.Namespace, domains)

		if err == nil {
			log.Debug().Msgf("Service %s.%s added in store", svc.Name, svc.Namespace)
		} else {
			log.Error().Err(err).Msgf("error adding service %s.%s in store", svc.Name, svc.Namespace)
		}
	}
}

func removeServiceFromStore(
	svc *corev1.Service,
	unlabelDeployment func(deployName, namespace string) error,
	deleteRouteFromStore func(id string) error,
) {
	if isAnnotationsProxlessCompatible(svc.ObjectMeta) {
		deployName := svc.Annotations[cluster.AnnotationServiceDeployKey]

		err := unlabelDeployment(deployName, svc.Namespace)

		if err != nil {
			log.Error().Err(err).Msgf("error unlabelling deployment %s.%s", deployName, svc.Namespace)
		}

		err = deleteRouteFromStore(string(svc.UID))

		if err == nil {
			log.Debug().Msgf("Service %s.%s removed from the store", svc.Name, svc.Namespace)
		} else {
			log.Error().Err(err).Msgf("error removing service %s.%s from store", svc.Name, svc.Namespace)
		}
	}
}

func updateServiceInStore(
	oldSvc, newSvc *corev1.Service,
	labelDeployment func(deployName, namespace string) error,
	unlabelDeployment func(deployName, namespace string) error,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	if isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) && isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // updating service
		newDeployName := newSvc.Annotations[cluster.AnnotationServiceDeployKey]
		oldDeployName := oldSvc.Annotations[cluster.AnnotationServiceDeployKey]

		// TODO add a test to only update the label if `oldDeployName` != `newDeployName`
		// /!\ if the deployment does not exist when the service is created, `deployName` will not be in the store
		// /!\ therefore, it will never be updated so the test has to be smarter than just a diff check.
		err := unlabelDeployment(oldDeployName, oldSvc.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("update - error unlabelling deployment %s.%s", oldDeployName, oldSvc.Namespace)
		}

		err = labelDeployment(newDeployName, newSvc.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("update - error labelling deployment %s.%s", newDeployName, newSvc.Namespace)
		}

		port := getPortFromServicePorts(newSvc.Spec.Ports)
		domains := genDomains(newSvc.Annotations[cluster.AnnotationServiceDomainKey], newSvc.Name, newSvc.Namespace)

		err = upsertStore(string(newSvc.UID), newSvc.Name, port, newDeployName, newSvc.Namespace, domains)
		if err == nil {
			log.Debug().Msgf("Service %s.%s updated in store", newSvc.Name, newSvc.Namespace)
		} else {
			log.Error().Err(err).Msgf("error updating service %s.%s in store", newSvc.Name, newSvc.Namespace)
		}
	} else if !isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) && isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // adding new service
		addServiceToStore(newSvc, labelDeployment, upsertStore)
	} else if isAnnotationsProxlessCompatible(oldSvc.ObjectMeta) && !isAnnotationsProxlessCompatible(newSvc.ObjectMeta) { // removing service
		removeServiceFromStore(oldSvc, unlabelDeployment, deleteRouteFromStore)
	}
}
