package controller

import (
	"github.com/stretchr/testify/assert"
	"kube-proxless/internal/cluster/fake"
	"kube-proxless/internal/config"
	"kube-proxless/internal/memory"
	"kube-proxless/internal/model"
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

func TestController_GetRouteByDomainFromMemory(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), nil, nil)

	// error - memory is empty
	_, err := c.GetRouteByDomainFromMemory("mock.io")
	assert.Error(t, err)

	route, err := model.NewRoute(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns",
		[]string{"mock.io"},
		nil, nil)
	assert.NoError(t, err)

	// add route in memory and test again
	assert.NoError(t, c.memory.UpsertMemoryMap(route))

	r, err := c.GetRouteByDomainFromMemory("mock.io")
	assert.NoError(t, err)

	if r.GetId() != "mock-id" || r.GetService() != "mock-svc" || r.GetDeployment() != "mock-deploy" ||
		r.GetNamespace() != "mock-ns" || !utils.CompareUnorderedArray(r.GetDomains(), []string{"mock.io"}) {
		t.Errorf("GetRouteByDomainFromMemory('mock.io') = %v; route in memory does not match", r)
	}
}

func TestController_UpdateLastUseMemory(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), nil, nil)

	// error - memory is empty
	assert.Error(t, c.UpdateLastUsedInMemory("mock.io"))

	route, err := model.NewRoute(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns",
		[]string{"mock.io"},
		nil, nil)
	assert.NoError(t, err)

	// add route in memory and test again
	assert.NoError(t, c.memory.UpsertMemoryMap(route))

	routeBefore, err := c.GetRouteByDomainFromMemory("mock.io")
	assert.NoError(t, err)
	timeBefore := routeBefore.GetLastUsed()

	time.Sleep(time.Second)

	assert.NoError(t, c.UpdateLastUsedInMemory("mock.io"))

	routeAfter, err := c.GetRouteByDomainFromMemory("mock.io")
	assert.NoError(t, err)

	if !routeAfter.GetLastUsed().After(timeBefore) {
		t.Errorf("UpdateLastUsedInMemory(); routeAfter = %s <= routeBefore = %s",
			routeAfter.GetLastUsed(), timeBefore)
	}
}

func TestController_ScaleUpDeployment(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), fake.NewCluster(), nil)

	// check the implemention of the fake client to understand the test

	assert.NoError(t, c.ScaleUpDeployment("mock-deploy", "mock-ns", 0))

	assert.Error(t, c.ScaleUpDeployment("deploy", "ns", 0))
}

func TestController_scaleDownDeployments(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), fake.NewCluster(), nil)

	helper_assertNoError(t, scaleDownDeployments(c))

	route, err := model.NewRoute(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns",
		[]string{"mock.io"},
		nil, nil)
	assert.NoError(t, err)

	// add route in memory and test again
	assert.NoError(t, c.memory.UpsertMemoryMap(route))

	helper_assertNoError(t, scaleDownDeployments(c))
}

func TestController_RunDownScaler(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), fake.NewCluster(), nil)

	// TODO check how we wanna deal with closing the channel and stopping the routine
	// We could use a context https://github.com/kubernetes/client-go/blob/master/examples/fake-client/main_test.go
	// but not sure if it is worth it
	go c.RunDownScaler(1)

	// make sure there is no panic when memory empty
	time.Sleep(1 * time.Second)

	route, err := model.NewRoute(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns",
		[]string{"mock.io"},
		nil, nil)
	assert.NoError(t, err)

	// add route in memorymemory and test again
	assert.NoError(t, c.memory.UpsertMemoryMap(route))

	// make sure there is no panic when memory had data
	time.Sleep(1 * time.Second)
}

func TestController_RunServicesEngine(t *testing.T) {
	c := NewController(memory.NewMemoryMap(), fake.NewCluster(), nil)

	// check the implemention of the fake client to understand the test
	config.NamespaceScope = "upsert"
	c.RunServicesEngine()

	_, err := c.memory.GetRouteByDomain("mock.io")
	assert.NoError(t, err)

	config.NamespaceScope = "delete"
	c.RunServicesEngine()

	_, err = c.memory.GetRouteByDomain("mock.io")
	assert.Error(t, err)

}
