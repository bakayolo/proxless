package controller

import "testing"

func helper_assertAtLeastOneError(t *testing.T, errs []error) {
	if errs == nil || len(errs) == 0 {
		t.Errorf("Array must have at least an error")
	}
}

func helper_assertNoError(t *testing.T, errs []error) {
	if errs != nil && len(errs) > 0 {
		t.Errorf("Array must not have any error; %s", errs)
	}
}
