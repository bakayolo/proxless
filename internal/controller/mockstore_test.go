package controller

import "kube-proxless/internal/model"

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
