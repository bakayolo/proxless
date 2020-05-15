package memory

import (
	"testing"
)

func upsertMemoryMapHelper(testCases []upsertTestCaseStruct, t *testing.T, s *MemoryMap) {
	for _, tc := range testCases {
		errGot := s.UpsertMemoryMap(tc.id, tc.svc, tc.port, tc.deploy, tc.ns, tc.domains)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("UpsertMemoryMap(%s) = %v; errWanted = %t", tc.id, errGot, tc.errWanted)
		}

		if errGot == nil {
			upsertMemoryMapGetRouteHelper(tc.id, tc.id, t, s)
			deploymentKey := genDeploymentKey(tc.deploy, tc.ns)
			upsertMemoryMapGetRouteHelper(tc.id, deploymentKey, t, s)
			for _, d := range tc.domains {
				upsertMemoryMapGetRouteHelper(tc.id, d, t, s)
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
