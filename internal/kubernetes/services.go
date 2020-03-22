package kubernetes

import (
	"fmt"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"kube-proxless/internal/config"
	"kube-proxless/internal/store"
	"strings"
	"time"
)

func StartServiceInformer() {
	opts := make([]informers.SharedInformerOption, 0)
	if config.Namespace != "" {
		opts = append(opts, informers.WithNamespace(config.Namespace))
	}
	infFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, 30*time.Second, opts...)

	serviceInformer := infFactory.Core().V1().Services().Informer()
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateRoutingMapObjects(obj, false)
			return
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateRoutingMapObjects(newObj, false)
			return
		},
		DeleteFunc: func(obj interface{}) {
			updateRoutingMapObjects(obj, true)
			return
		},
	}
	serviceInformer.AddEventHandler(eventHandler)

	log.Info().Msgf("Starting Services Informer")
	stopCh := make(chan struct{})
	defer close(stopCh)
	serviceInformer.Run(stopCh)
}

func updateRoutingMapObjects(obj interface{}, toDelete bool) {
	svc, ok := obj.(*core.Service)
	if !ok {
		log.Error().Msgf(fmt.Sprintf("Event for invalid object; got %T want *core.Service", obj))
		return
	}

	if metav1.HasAnnotation(svc.ObjectMeta, config.AnnotationDomainNameKey) &&
		metav1.HasAnnotation(svc.ObjectMeta, config.AnnotationSvcNameKey) {
		identifier := string(svc.UID)
		internalDomainName := fmt.Sprintf("%s.%s", svc.Annotations[config.AnnotationSvcNameKey], svc.Namespace)
		if toDelete {
			store.DeleteObjectInStore(identifier)
			log.Debug().Msgf("Service %s deleted", internalDomainName)
		} else {
			port := "80"
			if svc.Annotations[config.AnnotationSvcPortKey] != "" {
				port = svc.Annotations[config.AnnotationSvcPortKey]
			}
			label := svc.Labels[config.LabelProxlessSvc]
			domains := strings.Split(svc.Annotations[config.AnnotationDomainNameKey], ",")
			domains = append(domains, internalDomainName)                                      // add fqdn
			domains = append(domains, fmt.Sprintf("%s.svc.cluster.local", internalDomainName)) // add fqdn
			store.UpdateStore(identifier, internalDomainName, port, label, svc.Namespace, domains)
			log.Debug().Msgf("Service %s updated", internalDomainName)
		}
	}
}
