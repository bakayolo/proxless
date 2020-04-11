package servicesengine

import (
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"kube-proxless/internal/store"
)

func StartServiceInformer(namespace string) {
	infFactory := genInformerFactory(namespace)
	serviceInformer := infFactory.Core().V1().Services().Informer()
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc, err := parseService(obj)
			if err != nil {
				log.Error().Err(err).Msgf("Cannot process service in AddFunc handler")
				return
			}
			addServiceToStore(*svc)
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

			updateServiceInStore(*oldSvc, *newSvc)
			return
		},
		DeleteFunc: func(obj interface{}) {
			svc, err := parseService(obj)
			if err != nil {
				log.Error().Err(err).Msgf("Cannot process service in AddFunc handler")
				return
			}
			removeServiceFromStore(*svc)
			return
		},
	}
	serviceInformer.AddEventHandler(eventHandler)

	log.Info().Msgf("Starting Services Informer")
	stopCh := make(chan struct{})
	defer close(stopCh)
	serviceInformer.Run(stopCh)
}

func addServiceToStore(svc core.Service) {
	if isProxlessCompatible(svc) {
		deployName := svc.Annotations[annotationDeployKey]

		addProxyLabelToDeployment(deployName, svc.Namespace)

		port := genPort(svc.Spec.Ports)
		domains := genDomains(svc.Annotations[annotationDomainKey], svc.Name, svc.Namespace)

		store.UpdateStore(stringifyUid(svc.UID), svc.Name, port, deployName, svc.Namespace, domains)
		log.Debug().Msgf("Service %s.%s added in store", svc.Name, svc.Namespace)
	}
}

func removeServiceFromStore(svc core.Service) {
	if isProxlessCompatible(svc) {
		deployName := svc.Annotations[annotationDeployKey]
		removeProxyLabelFromDeployment(deployName, svc.Namespace)

		store.DeleteObjectInStore(stringifyUid(svc.UID))
		log.Debug().Msgf("Service %s.%s deleted from store", svc.Name, svc.Namespace)
	}
}

func updateServiceInStore(oldSvc, newSvc core.Service) {
	if isProxlessCompatible(oldSvc) && isProxlessCompatible(newSvc) { // updating service
		newDeployName := newSvc.Annotations[annotationDeployKey]
		oldDeployName := oldSvc.Annotations[annotationDeployKey]

		// TODO add a test to only update the label if `oldDeployName` != `newDeployName`
		// /!\ if the deployment does not exist when the service is created, `deployName` will not in the store
		// /!\ therefore, it will never be update so the test has to be smarter than just a diff check.
		removeProxyLabelFromDeployment(oldDeployName, newSvc.Namespace)
		addProxyLabelToDeployment(newDeployName, newSvc.Namespace)

		port := genPort(newSvc.Spec.Ports)
		domains := genDomains(newSvc.Annotations[annotationDomainKey], newSvc.Name, newSvc.Namespace)

		store.UpdateStore(stringifyUid(newSvc.UID), newSvc.Name, port, newDeployName, newSvc.Namespace, domains)
		log.Debug().Msgf("Service %s.%s updated in store", newSvc.Name, newSvc.Namespace)
	} else if !isProxlessCompatible(oldSvc) && isProxlessCompatible(newSvc) { // adding new service
		addServiceToStore(newSvc)
	} else if isProxlessCompatible(oldSvc) && !isProxlessCompatible(newSvc) { // removing service
		removeServiceFromStore(oldSvc)
	}
}

func updateDeploymentProxyLabel(oldName, newName, namespace string) {
	if oldName != newName {

	}
}
