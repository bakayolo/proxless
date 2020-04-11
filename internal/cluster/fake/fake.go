package fake

import (
	"errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"kube-proxless/internal/logger"
	"time"
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

func NewCluster() *fakeCluster {
	return &fakeCluster{}
}

func (*fakeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	if name != deployName || namespace != namespaceName {
		return errors.New("error scaling up deployment")
	}
	return nil
}

func (*fakeCluster) ScaleDownDeployments(
	namespaceScope string, mustScaleDown func(deployName, namespace string) (bool, time.Duration, error)) []error {
	_, _, err := mustScaleDown(deployName, namespaceName)

	if err != nil {
		return append([]error{}, err)
	}

	return nil
}

func (*fakeCluster) RunServicesEngine(
	namespaceScope, proxlessService, proxlessNamespace string,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	if namespaceScope == "upsert" { // TODO this is too hacky, see how others are doing
		err := upsertStore(
			serviceId, serviceName, "", deployName, namespaceName, domains)

		if err != nil {
			logger.Errorf(err, "Error upserting in fake package")
		}
	} else {
		err := deleteRouteFromStore(serviceId)

		if err != nil {
			logger.Errorf(err, "Error deleting route in fake package")
		}
	}
}
