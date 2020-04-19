package servicesengine

import (
	"kube-proxless/internal/model"
	"time"
)

type mockCluster struct{}

func (*mockCluster) ScaleUpDeployment(name, namespace string) error {
	return nil
}

func (*mockCluster) RunDownScaler(namespace string, timeout time.Duration, mustScaleDown func(name, namespace string) bool) {
}

func (*mockCluster) RunServicesEngine(
	namespace string,
	labelDeployment func(deployName, namespace string) error,
	unlabelDeployment func(deployName, namespace string) error,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
}

func (*mockCluster) LabelDeployment(name, namespace string) error {
	return nil
}

func (*mockCluster) UnlabelDeployment(name, namespace string) error {
	return nil
}

type mockStore struct{}

func (*mockStore) UpsertStore(id, service, port, deploy, namespace string, domains []string) error {
	return nil
}

func (*mockStore) GetRouteByDomain(domain string) (*model.Route, error) {
	return model.NewRoute("mock", "mock", "", "mock", "mock", []string{"mock.io"})
}

func (*mockStore) GetRouteByDeployment(deploy, namespace string) (*model.Route, error) {
	return model.NewRoute("mock", "mock", "", "mock", "mock", []string{"mock.io"})
}

func (*mockStore) UpdateLastUse(domain string) error {
	return nil
}

func (*mockStore) DeleteRoute(id string) error {
	return nil
}
