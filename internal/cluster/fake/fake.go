package fake

import (
	"errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/logger"
)

const (
	deployName    = "mock-deploy"
	namespaceName = "mock-ns"
	serviceId     = "mock-id"
	serviceName   = "mock-svc"
)

var (
	domains = []string{"mock.io"}
)

type fakeCluster struct{}

func NewCluster() cluster.Interface {
	return &fakeCluster{}
}

func (*fakeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	if name != deployName || namespace != namespaceName {
		return errors.New("error scaling up deployment")
	}
	return nil
}

func (*fakeCluster) ScaleDownDeployment(deploymentName, namespace string) error {
	return nil
}

func (*fakeCluster) RunServicesEngine(
	namespaceScope, proxlessService, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, isRunning bool, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	if namespaceScope == "upsert" { // TODO this is too hacky, see how others are doing
		err := upsertMemory(
			serviceId, serviceName, "", deployName, namespaceName, domains, true, nil, nil)

		if err != nil {
			logger.Errorf(err, "Error upserting in fake package")
		}
	} else {
		err := deleteRouteFromMemory(serviceId)

		if err != nil {
			logger.Errorf(err, "Error deleting route in fake package")
		}
	}
}
