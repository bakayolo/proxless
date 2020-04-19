package cluster

import (
	"time"
)

type ClusterInterface interface {
	ScaleUpDeployment(name, namespace string) error

	RunDownScaler(namespace string, checkInterval time.Duration, mustScaleDown func(name, namespace string) bool)

	RunServicesEngine(
		namespace string,
		labelDeployment func(deployName, namespace string) error,
		unlabelDeployment func(deployName, namespace string) error,
		upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
		deleteRouteFromStore func(id string) error,
	)
	LabelDeployment(name, namespace string) error
	UnlabelDeployment(name, namespace string) error
}
