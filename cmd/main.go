package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	ds "kube-proxless/internal/downscaler"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/server/http"
	se "kube-proxless/internal/servicesengine"
	"kube-proxless/internal/store/inmemory"
)

func main() {
	logger.InitLogger()

	config.LoadEnvVars()

	store := inmemory.NewInMemoryStore()
	cluster := kube.NewKubeClient()

	downScaler := ds.NewDownScaler(store, cluster)
	go downScaler.Run()

	serviceEngine := se.NewServicesEngine(store, cluster)
	go serviceEngine.Run()

	controller := ctrl.NewController(store, cluster)
	httpServer := http.NewHTTPServer(controller)
	httpServer.Run()
}
