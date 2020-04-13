package controller

import (
	"kube-proxless/internal/model"
	"kube-proxless/internal/store"
)

type ControllerInterface interface {
	GetRouteByDomainFromStore(domain string) (*model.Route, error)
	UpdateLastUseInStore(domain string) error
}

type Controller struct {
	store store.StoreInterface
}

func NewController(store store.StoreInterface) *Controller {
	return &Controller{
		store: store,
	}
}

func (c *Controller) GetRouteByDomainFromStore(domain string) (*model.Route, error) {
	return c.store.GetRouteByDomain(domain)
}

func (c *Controller) UpdateLastUseInStore(domain string) error {
	return c.store.UpdateLastUse(domain)
}
