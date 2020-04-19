package cluster

import "time"

type ClusterInterface interface {
	ScaleUpDeployment(name, namespace string) error

	RunDownScaler(namespace string, checkInterval time.Duration, mustScaleDown func(name, namespace string) bool)
}
