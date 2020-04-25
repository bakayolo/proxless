package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/logger"
	"os"
	"strconv"
)

var (
	KubeConfigPath             string
	Port                       string
	MaxConsPerHost             int
	Namespace                  string
	ServerlessTTL              int
	DeploymentReadinessTimeout int
)

func LoadEnvVars() {
	KubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")

	Port = getString("PORT", "80")
	MaxConsPerHost = getInt("MAX_CONS_PER_HOST", 10000)

	Namespace = os.Getenv("NAMESPACE")

	ServerlessTTL = getInt("SERVERLESS_TTL_SECONDS", 30)
	DeploymentReadinessTimeout = getInt("DEPLOYMENT_READINESS_TIMEOUT_SECONDS", 30)
}

func getString(key, fallback string) string {
	var result string
	if os.Getenv(key) != "" {
		result = os.Getenv(key)
	} else {
		result = fallback
	}

	logger.Debugf("Successfully loaded env var: %s=%v", key, result)

	return result
}

func getInt(key string, fallback int) int {
	var result int
	if os.Getenv(key) != "" {
		intVal, err := strconv.Atoi(os.Getenv(key))
		if err != nil {
			log.Panic().Err(err).Msgf("error parsing int from env var: %s", key)
		}
		result = intVal
	} else {
		result = fallback
	}

	log.Debug().Msgf("Successfully loaded env var: %s=%v", key, result)

	return result
}
