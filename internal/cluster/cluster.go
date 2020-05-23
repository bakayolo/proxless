package cluster

import "time"

type Interface interface {
	ScaleUpDeployment(name, namespace string, timeout int) error

	ScaleDownDeployments(
		namespaceScope string, mustScaleDown func(name, namespace string) (bool, time.Duration, error)) []error

	RunServicesEngine(
		namespaceScope, proxlessService, proxlessNamespace string,
		upsertMemory func(
			id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
		deleteRouteFromMemory func(id string) error,
	)
}
