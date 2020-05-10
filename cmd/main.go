package main

import (
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/memory"
	"kube-proxless/internal/pubsub"
	"kube-proxless/internal/pubsub/redis"
	"kube-proxless/internal/server/http"
)

func main() {
	logger.InitLogger()

	config.LoadEnvVars()

	memoryMap := memory.NewMemoryMap()
	cluster := kube.NewCluster(kube.NewKubeClient(config.KubeConfigPath))

	var ps pubsub.Interface
	if config.RedisURL != "" {
		ps = redis.NewRedisPubSub(config.RedisURL)
	}

	controller := ctrl.NewController(memoryMap, cluster, ps)

	go controller.RunDownScaler(30) // TODO make `checkInterval` configurable

	go controller.RunServicesEngine()

	http.NewHTTPServer(controller).Run()
}
