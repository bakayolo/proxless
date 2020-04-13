package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	"kube-proxless/internal/kubernetes"
	"kube-proxless/internal/kubernetes/downscaler"
	"kube-proxless/internal/kubernetes/servicesengine"
	"kube-proxless/internal/server/http"
	"kube-proxless/internal/store/inmemory"
)

func main() {
	config.LoadConfig()
	log.Info().Msgf("Log Level is %s", config.InitLogger())

	kubernetes.InitKubeClient()
	go servicesengine.StartServiceInformer(config.Namespace)
	go downscaler.StartDownScaler()

	store := inmemory.NewInMemoryStore()
	cluster := kube.NewKubeClient()

	controller := ctrl.NewController(store, cluster)

	httpServer := http.NewHTTPServer(controller)
	httpServer.StartServer()
}
