package kube

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"kube-proxless/internal/logger"
	"time"
)

func runServicesInformer(
	clientSet kubernetes.Interface,
	namespaceScope, proxlessService, proxlessNamespace string,
	informerResyncInterval int,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	namespaceScoped := false
	opts := make([]informers.SharedInformerOption, 0)
	if namespaceScope != "" {
		opts = append(opts, informers.WithNamespace(namespaceScope))
		namespaceScoped = true
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

			addServiceToMemory(clientSet, svc, namespaceScoped, proxlessService, proxlessNamespace, upsertMemory)

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

			updateServiceMemory(
				clientSet, oldSvc, newSvc, namespaceScoped, proxlessService, proxlessNamespace,
				upsertMemory, deleteRouteFromMemory)
			return
		},
		DeleteFunc: func(obj interface{}) {
			svc, err := parseService(obj)

			if err != nil {
				logger.Errorf(err, "Cannot process service in DeleteFunc handler")
				return
			}

			removeServiceFromMemory(clientSet, svc, deleteRouteFromMemory)

			return
		},
	}
	informer.AddEventHandler(eventHandler)

	stopCh := make(chan struct{})
	defer close(stopCh)
	informer.Run(stopCh)
}
