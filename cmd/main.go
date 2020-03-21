package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes"
	"kube-proxless/internal/server"
)

func main() {
	config.LoadConfig()
	log.Info().Msgf("Log Level is %s", config.InitLogger())

	go kubernetes.StartServiceInformer()

	server.StartServer()
}
