package memory

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"kube-proxless/internal/model"
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

// we volontarily do not create the MemoryMap globally so that each test are independent from each other

type upsertTestCaseStruct struct {
	id, svc, port, deploy, ns string
	domains                   []string
	errWanted                 bool
}

func TestMemoryMap_UpsertMap_Create(t *testing.T) {
	s := NewMemoryMap()

	// create route
	testCases := []upsertTestCaseStruct{
		{"createTestCase0", "svc0", "80", "deploy0", "ns0", []string{"example.0"}, false},
		{"createTestCase1", "svc1", "", "deploy1", "ns1", []string{"example.1"}, false},
	}

	upsertMemoryMapHelper(testCases, t, s)
}

func TestMemoryMap_UpsertMemoryMap_Update(t *testing.T) {
	s := NewMemoryMap()

	testCases := []upsertTestCaseStruct{
		{"updateTestCase0", "svc0", "80", "deploy0", "ns", []string{"example.0.0"}, false},
		{"updateTestCase0", "svc0", "80", "deploy0", "ns", []string{"example.0.0"}, false},
		{"updateTestCase0", "svc0", "80", "deploy0.1", "ns", []string{"example.0.0", "example.0.1"}, false},
		{"updateTestCase1", "svc1", "", "deploy1", "ns1", []string{"example.1.0"}, false},
		{"updateTestCase1", "svc1", "8080", "deploy1", "ns1", []string{"example.1.0"}, false},
		{"updateTestCase1", "svc1", "", "deploy1", "ns1", []string{"example.1.0"}, false},
		{"updateTestCase2", "svc2", "8080", "deploy2", "ns1", []string{"example.1.0"}, true},
		{"updateTestCase2", "svc2", "8080", "deploy2", "ns1", []string{"example.1.1"}, false},
	}

	upsertMemoryMapHelper(testCases, t, s)
}

func TestMemoryMap_genDeploymentKey(t *testing.T) {
	deploy := "exampledeploy"
	ns := "examplens"
	want := fmt.Sprintf("%s.%s", deploy, ns)

	deploymentKey := genDeploymentKey(deploy, ns)

	if deploymentKey != want {
		t.Errorf("genDeploymentKey(%s, %s) = %s; want = %s", deploy, ns, deploymentKey, want)
	}
}

func TestMemoryMap_CheckDeployAndDomainsOwnership(t *testing.T) {
	s := NewMemoryMap()

	route, err := model.NewRoute(
		"0", "svc0", "", "deploy0", "ns0", []string{"example.0.0"}, true, nil, nil)
	assert.NoError(t, err)

	err = s.UpsertMemoryMap(route)
	assert.NoError(t, err)

	testCases := []struct {
		id, deploy, ns string
		domains        []string
		errWanted      bool
	}{
		{"0", "deploy0", "ns0", []string{"example.0.0"}, false},
		{"0", "deploy0", "ns0", []string{"example.0.0"}, false},
		{"0", "deploy0", "ns0", []string{"example.0.1"}, false},
		{"1", "deploy1", "ns1", []string{"example.0.0"}, true},
		{"1", "deploy0", "ns0", []string{"example.1.0"}, true},
	}

	for _, tc := range testCases {
		errGot := checkDeployAndDomainsOwnership(s, tc.id, tc.deploy, tc.ns, tc.domains)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("checkDeployAndDomainsOwnership(%s, %s ,%s, %s) = %v, errWanted = %t",
				tc.id, tc.deploy, tc.ns, tc.domains, errGot, tc.errWanted)
		}
	}
}

func TestMemoryMap_cleanOldDeploymentFromMap(t *testing.T) {
	s := NewMemoryMap()

	r0, err :=
		model.NewRoute("0", "svc0", "", "deploy0", "ns0", []string{"example.0.0"}, true, nil, nil)
	assert.NoError(t, err)

	createRoute(s, r0)

	testCases := []struct {
		id      string
		route   *model.Route
		domains []string
		want    []string
	}{
		{"0", r0, r0.GetDomains(), []string{}},
		{"1", r0, []string{"example.0.0", "example.0.1"}, []string{"example.0.1"}},
		{"2", r0, []string{"example.0.0", "example.0.1"}, []string{"example.0.1"}},
		{"3", r0, []string{"example.0.1"}, []string{"example.0.1"}},
		{"4", r0, []string{"example.0.0"}, []string{}}, // the map has been updated but the route did not change
	}

	for _, tc := range testCases {
		got := cleanOldDomainsFromMemoryMap(s, tc.route.GetDomains(), tc.domains)

		if !utils.CompareUnorderedArray(got, tc.want) {
			t.Errorf("cleanOldDeploymentFromMemoryMap(id = %s, %s) = %s; want = %s", tc.id, tc.domains, got, tc.want)
		}
	}
}

func TestMemoryMap_UpdateLastUse(t *testing.T) {
	s := NewMemoryMap()

	r0, err :=
		model.NewRoute("0", "svc0", "", "deploy0", "ns0", []string{"example.0.0"}, true, nil, nil)
	assert.NoError(t, err)

	createRoute(s, r0)

	lastUsed := r0.GetLastUsed()

	testCases := []struct {
		id        string
		errWanted bool
	}{
		{r0.GetId(), false},
		{"", true},
	}

	for _, tc := range testCases {
		errGot := s.UpdateLastUsed(tc.id, time.Now())

		if tc.errWanted != (errGot != nil) {
			t.Errorf("UpdateLastUsed(%s) = %v; errWanted = %t", tc.id, errGot, tc.errWanted)
		}

		if errGot == nil && !lastUsed.Before(r0.GetLastUsed()) {
			t.Errorf("UpdateLastUsed(%s) - %s is not before %s", tc.id, lastUsed, r0.GetLastUsed())
		}
	}
}

func TestMemoryMap_DeleteRoute(t *testing.T) {
	s := NewMemoryMap()

	r0, _ := model.NewRoute("0", "svc0", "", "deploy0", "ns0", []string{"example.0.0"}, true, nil, nil)
	createRoute(s, r0)

	testCases := []struct {
		id        string
		errWanted bool
	}{
		{r0.GetId(), false},
		{r0.GetId(), true},
	}

	for _, tc := range testCases {
		errGot := s.DeleteRoute(tc.id)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("DeleteRoute(%s) = %v; errWanted = %t", tc.id, errGot, tc.errWanted)
		}

		if errGot == nil {
			_, err := getRoute(s, tc.id)

			if err == nil {
				t.Errorf("DeleteRoute(%s) = %v; route still in memory", tc.id, errGot)
			}
		}
	}
}
