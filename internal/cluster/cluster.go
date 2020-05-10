package cluster

import "time"

type Interface interface {
	ScaleUpDeployment(name, namespace string, timeout int) error

	ScaleDownDeployments(
		namespaceScope string, mustScaleDown func(name, namespace string) (bool, time.Duration, error)) []error

	RunServicesEngine(
		namespaceScope, proxlessService, proxlessNamespace string,
		upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
		deleteRouteFromStore func(id string) error,
	)
}
