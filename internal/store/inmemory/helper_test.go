package inmemory

import (
	"testing"
)

func upsertStoreHelper(testCases []upsertTestCaseStruct, t *testing.T, s *inMemoryStore) {
	for _, tc := range testCases {
		errGot := s.UpsertStore(tc.id, tc.svc, tc.port, tc.deploy, tc.ns, tc.domains)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("UpsertStore(%s) = %v; errWanted = %t", tc.id, errGot, tc.errWanted)
		}

		if errGot == nil {
			upsertStoreGetRouteHelper(tc.id, tc.id, t, s)
			deploymentKey := genDeploymentKey(tc.deploy, tc.ns)
			upsertStoreGetRouteHelper(tc.id, deploymentKey, t, s)
			for _, d := range tc.domains {
				upsertStoreGetRouteHelper(tc.id, d, t, s)
			}
		}
	}
}

func upsertStoreGetRouteHelper(id, key string, t *testing.T, s *inMemoryStore) {
	_, err := getRoute(s, key)

	if err != nil {
		t.Errorf("UpsertStore(%s); key %s not in s", id, key)
	}
}
