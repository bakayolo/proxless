package main

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/config"
	"kube-proxless/internal/server"
)

func main() {
	config.LoadConfig()
	log.Info().Msgf("Log Level is %s", config.InitLogger())

	server.StartServer()
}
