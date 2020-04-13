package cluster

type ClusterInterface interface {
	ScaleUpDeployment(name, namespace string) error
}
