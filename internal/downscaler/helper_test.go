package downscaler

import (
	"errors"
	"kube-proxless/internal/model"
	"time"
)

type mockCluster struct{}

func (*mockCluster) ScaleUpDeployment(name, namespace string) error {
	return nil
}

func (*mockCluster) RunDownScaler(namespace string, timeout time.Duration, mustScaleDown func(name, namespace string) bool) {
}

type mockStore struct{}

func (*mockStore) UpsertStore(id, service, port, deploy, namespace string, domains []string) error {
	return nil
}

func (*mockStore) GetRouteByDomain(domain string) (*model.Route, error) {
	return nil, nil
}

func (*mockStore) GetRouteByDeployment(deploy, namespace string) (*model.Route, error) {
	if deploy == "error" {
		return nil, errors.New("route not found")
	}

	route, _ := model.NewRoute("mock", "mock", "", "mock", "mock", []string{"mock.io"})

	if deploy == "timeout" {
		route.SetLastUsed(time.Now().AddDate(-1, 0, 0)) // minus 1 year
	}

	if deploy == "notimeout" {
		route.SetLastUsed(time.Now().AddDate(1, 0, 0)) // add 1 year
	}

	return route, nil
}

func (*mockStore) UpdateLastUse(domain string) error {
	return nil
}

func (*mockStore) DeleteRoute(id string) error {
	return nil
}
