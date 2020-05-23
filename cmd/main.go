package main

import (
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/cluster/kube"
	"kube-proxless/internal/cluster/openshift"
	"kube-proxless/internal/config"
	ctrl "kube-proxless/internal/controller"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/memory"
	"kube-proxless/internal/pubsub"
	"kube-proxless/internal/pubsub/redis"
	"kube-proxless/internal/server/http"
	"strings"
)

func main() {
	logger.InitLogger()

	config.LoadEnvVars()

	memoryMap := memory.NewMemoryMap()

	servicesInformerResyncInterval := 60
	var c cluster.Interface

	if strings.ToUpper(config.Cluster) == "OPENSHIFT" {
		c = openshift.NewCluster(
			kube.NewKubeClient(config.KubeConfigPath),
			openshift.NewOpenshiftClient(config.KubeConfigPath),
			servicesInformerResyncInterval)
	} else {
		c = kube.NewCluster(kube.NewKubeClient(config.KubeConfigPath), servicesInformerResyncInterval)
	}

	var ps pubsub.Interface
	if config.RedisURL != "" {
		ps = redis.NewRedisPubSub(config.RedisURL)
	}

	controller := ctrl.NewController(memoryMap, c, ps)

	go controller.RunDownScaler(30) // TODO make `checkInterval` configurable

	go controller.RunServicesEngine()

	http.NewHTTPServer(controller).Run()
}
