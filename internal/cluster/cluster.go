package cluster

type Interface interface {
	ScaleUpDeployment(name, namespace string, timeout int) error

	ScaleDownDeployment(deploymentName, namespace string) error

	RunServicesEngine(
		namespaceScope, proxlessService, proxlessNamespace string,
		upsertMemory func(
			id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
		deleteRouteFromMemory func(id string) error,
	)
}
