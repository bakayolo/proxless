package kubernetes

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/commons"
	"kube-proxless/internal/config"
	"time"
)

var (
	annotationKey = "proxless/domain-name"
)

func StartServiceInformer() {
	kubeClient := getKubeClient()

	opts := make([]informers.SharedInformerOption, 0)
	if config.Namespace != "" {
		opts = append(opts, informers.WithNamespace(config.Namespace))
	}
	infFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 30*time.Second, opts...)

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

	if metav1.HasAnnotation(svc.ObjectMeta, annotationKey) {
		internalDomainName := fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)
		if toDelete {
			commons.DeleteRoute(internalDomainName, svc.Annotations[annotationKey])
			log.Debug().Msgf("Service %s deleted", internalDomainName)
		} else {
			commons.UpdateRoute(internalDomainName, internalDomainName)
			commons.UpdateRoute(svc.Annotations[annotationKey], internalDomainName)
			log.Debug().Msgf("Service %s updated", internalDomainName)
		}
	}
}

func getKubeClient() *kubernetes.Clientset {
	kubeConf := loadKubeConfig(config.KubeConfigPath)
	return kubernetes.NewForConfigOrDie(kubeConf)
}

func loadKubeConfig(kubeConfigPath string) *rest.Config {
	kubeConfigString := flag.String("kubeconfig", kubeConfigPath, "(optional) absolute path to the kubeconfig file")

	// use the current context in kubeconfig
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeConfigString)
	if err != nil {
		log.Panic().Err(err).Msgf("Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return kubeConfig
}
