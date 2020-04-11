package store

import "kube-proxless/internal/model"

type StoreInterface interface {
	UpsertStore(id, service, port, deploy, namespace string, domains []string) error
	GetRouteByDomain(domain string) (*model.Route, error)
	GetRouteByDeployment(deploy, namespace string) (*model.Route, error)
	UpdateLastUse(domain string) error
	DeleteRoute(id string) error
}
