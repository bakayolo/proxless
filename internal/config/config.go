package config

import (
	_ "github.com/joho/godotenv/autoload"
	"kube-proxless/internal/logger"
	"os"
	"strconv"
)

var (
	KubeConfigPath                    string
	Port                              string
	MaxConsPerHost                    int
	ProxlessNamespace                 string
	ProxlessService                   string
	NamespaceScope                    string
	ServerlessTTLSeconds              int
	DeploymentReadinessTimeoutSeconds int
)

func LoadEnvVars() {
	KubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")

	Port = getString("PORT", "80")
	MaxConsPerHost = getInt("MAX_CONS_PER_HOST", 10000)

	ProxlessNamespace = getString("PROXLESS_NAMESPACE", "proxless")
	ProxlessService = getString("PROXLESS_SERVICE", "proxless")

	if getBool("NAMESPACE_SCOPED", true) {
		NamespaceScope = ProxlessNamespace
	}

	ServerlessTTLSeconds = getInt("SERVERLESS_TTL_SECONDS", 30)
	DeploymentReadinessTimeoutSeconds = getInt("DEPLOYMENT_READINESS_TIMEOUT_SECONDS", 30)
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
			logger.Panicf(err, "error parsing int from env var: %s", key)
		}
		result = intVal
	} else {
		result = fallback
	}

	logger.Debugf("Successfully loaded env var: %s=%v", key, result)

	return result
}

func getBool(key string, fallback bool) bool {
	var result bool
	if os.Getenv(key) != "" {
		boolVal, err := strconv.ParseBool(os.Getenv(key))
		if err != nil {
			logger.Panicf(err, "error parsing bool from env var: %s", key)
		}
		result = boolVal
	} else {
		result = fallback
	}

	logger.Debugf("Successfully loaded env var: %s=%v", key, result)

	return result
}
