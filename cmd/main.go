package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes"
	"kube-proxless/internal/kubernetes/downscaler"
	"kube-proxless/internal/kubernetes/servicesfactory"
	"kube-proxless/internal/server"
)

func main() {
	config.LoadConfig()
	log.Info().Msgf("Log Level is %s", config.InitLogger())

	kubernetes.LoadKubeClient()
	go servicesfactory.StartServiceInformer(config.Namespace)
	go downscaler.StartDownScaler()

	server.StartServer()
}
