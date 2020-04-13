package controller

import (
	"testing"
)

var (
	c = NewController(&mockStore{})
)

func TestController_GetRouteByDomainFromStore(t *testing.T) {
	r, err := c.GetRouteByDomainFromStore("")

	if err != nil {
		t.Errorf("GetRouteByDomainFromStore() = (%v, %v); must not error", r, err)
	}
}

func TestController_UpdateLastUseInStore(t *testing.T) {
	err := c.UpdateLastUseInStore("")

	if err != nil {
		t.Errorf("UpdateLastUseInStore() = %v; must not error", err)
	}
}
