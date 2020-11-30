package memory

import (
	"github.com/stretchr/testify/assert"
	"kube-proxless/internal/model"
	"testing"
)

func upsertMemoryMapHelper(testCases []upsertTestCaseStruct, t *testing.T, s *MemoryMap) {
	for _, tc := range testCases {
		route, err := model.NewRoute(
			tc.id, tc.svc, tc.port, tc.deploy, tc.ns, tc.domains, true, nil, nil)
		assert.NoError(t, err)

		if err == nil {
			err = s.UpsertMemoryMap(route)

			if tc.errWanted != (err != nil) {
				t.Errorf("UpsertMemoryMap(%s) = %v; errWanted = %t", tc.id, err, tc.errWanted)
			}

			if err == nil {
				upsertMemoryMapGetRouteHelper(tc.id, tc.id, t, s)
				deploymentKey := genDeploymentKey(tc.deploy, tc.ns)
				upsertMemoryMapGetRouteHelper(tc.id, deploymentKey, t, s)
				for _, d := range tc.domains {
					upsertMemoryMapGetRouteHelper(tc.id, d, t, s)
				}
			}
		}
	}
}

func upsertMemoryMapGetRouteHelper(id, key string, t *testing.T, s *MemoryMap) {
	_, err := getRoute(s, key)

	if err != nil {
		t.Errorf("UpsertMemoryMap(%s); key %s not in s", id, key)
	}
}
