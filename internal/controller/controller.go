package controller

import (
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/model"
	"kube-proxless/internal/store"
)

type ControllerInterface interface {
	GetRouteByDomainFromStore(domain string) (*model.Route, error)
	UpdateLastUseInStore(domain string) error

	ScaleUpDeployment(name, namespace string) error
}

type Controller struct {
	store   store.StoreInterface
	cluster cluster.ClusterInterface
}

func NewController(store store.StoreInterface, cluster cluster.ClusterInterface) *Controller {
	return &Controller{
		store:   store,
		cluster: cluster,
	}
}

func (c *Controller) GetRouteByDomainFromStore(domain string) (*model.Route, error) {
	return c.store.GetRouteByDomain(domain)
}

func (c *Controller) UpdateLastUseInStore(domain string) error {
	return c.store.UpdateLastUse(domain)
}

func (c *Controller) ScaleUpDeployment(name, namespace string) error {
	return c.cluster.ScaleUpDeployment(name, namespace)
}
