package controller

import (
	"testing"
)

var (
	c = NewController(&mockStore{}, &mockCluster{})
)

func TestController_GetRouteByDomainFromStore(t *testing.T) {
	r, err := c.GetRouteByDomainFromStore("")

	if err != nil {
		t.Errorf("GetRouteByDomainFromStore() = (%v, %v); error must be nil", r, err)
	}
}

func TestController_UpdateLastUseInStore(t *testing.T) {
	err := c.UpdateLastUseInStore("")

	if err != nil {
		t.Errorf("UpdateLastUseInStore() = %v; error must be nil", err)
	}
}

func TestController_ScaleUpDeployment(t *testing.T) {
	err := c.ScaleUpDeployment("", "")

	if err != nil {
		t.Errorf("ScaleUpDeployment() = %v; error must be nil", err)
	}
}
