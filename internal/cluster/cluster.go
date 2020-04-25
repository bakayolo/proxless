package cluster

type ClusterInterface interface {
	ScaleUpDeployment(name, namespace string, timeout int) error

	ScaleDownDeployments(namespace string, mustScaleDown func(name, namespace string) bool) []error

	RunServicesEngine(
		namespace string,
		upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
		deleteRouteFromStore func(id string) error,
	)
}
