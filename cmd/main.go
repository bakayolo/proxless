package main

import (
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/server/http"
	"kube-proxless/internal/store/inmemory"
)

func main() {
	logger.InitLogger()

	config.LoadEnvVars()

	store := inmemory.NewInMemoryStore()
	cluster := kube.NewCluster(kube.NewKubeClient(config.KubeConfigPath))

	controller := ctrl.NewController(store, cluster)

	go controller.RunDownScaler(30) // TODO make `checkInterval` configurable

	go controller.RunServicesEngine()

	http.NewHTTPServer(controller).Run()
}
