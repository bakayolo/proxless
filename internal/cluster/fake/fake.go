package fake

import (
	"errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type fakeCluster struct{}

func NewCluster() *fakeCluster {
	return &fakeCluster{}
}

func (c *fakeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	if name != "mock-deploy" || namespace != "mock-ns" {
		return errors.New("error scaling up deployment")
	}
	return nil
}

func (c *fakeCluster) ScaleDownDeployments(
	namespace string, mustScaleDown func(deployName, namespace string) (bool, error)) []error {
	_, err := mustScaleDown("mock-deploy", "mock-ns")

	if err != nil {
		return append([]error{}, err)
	}

	return nil
}

func (c *fakeCluster) RunServicesEngine(
	namespace string,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	if namespace == "upsert" {
		_ = upsertStore(
			"mock-id", "mock-svc", "", "mock-deploy", "mock-ns", []string{"mock.io"})
	} else {
		_ = deleteRouteFromStore("mock-id")
	}
}
