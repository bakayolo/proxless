package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	ds "kube-proxless/internal/downscaler"
	"kube-proxless/internal/kubernetes/servicesengine"
	"kube-proxless/internal/server/http"
	"kube-proxless/internal/store/inmemory"
)

func main() {
	config.LoadEnvVars()
	log.Info().Msgf("Log Level is %s", config.InitLogger())

	go servicesengine.StartServiceInformer(config.Namespace)

	store := inmemory.NewInMemoryStore()
	cluster := kube.NewKubeClient()

	downScaler := ds.NewDownScaler(store, cluster)
	go downScaler.Run()

	controller := ctrl.NewController(store, cluster)
	httpServer := http.NewHTTPServer(controller)
	httpServer.Run()
}
