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
	"time"
)

var (
	annotationDomainNameKey = "proxless/domain-name"
	annotationSvcNameKey    = "proxless/service-name"
	annotationSvcPortKey    = "proxless/service-port"

	labelSvc = "proxless"
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
	serviceInformer.Run(stopCh)
}

func updateRoutingMapObjects(obj interface{}, toDelete bool) {
	svc, ok := obj.(*core.Service)
	if !ok {
		log.Error().Msgf(fmt.Sprintf("Event for invalid object; got %T want *core.Service", obj))
		return
	}

	if metav1.HasAnnotation(svc.ObjectMeta, annotationDomainNameKey) &&
		metav1.HasAnnotation(svc.ObjectMeta, annotationSvcNameKey) {
		internalDomainName := fmt.Sprintf("%s.%s", svc.Annotations[annotationSvcNameKey], svc.Namespace)
		if toDelete {
			store.DeleteRoute(internalDomainName, svc.Annotations[annotationDomainNameKey])
			log.Debug().Msgf("Service %s deleted", internalDomainName)
		} else {
			store.UpdateRoute(internalDomainName, internalDomainName, svc.Annotations[annotationSvcPortKey], svc.Labels[labelSvc], svc.Namespace)
			store.UpdateRoute(svc.Annotations[annotationDomainNameKey], internalDomainName, svc.Annotations[annotationSvcPortKey], svc.Labels[labelSvc], svc.Namespace)
			log.Debug().Msgf("Service %s updated", internalDomainName)
		}
	}
}
