package downscaler

import "testing"

var (
	ds = NewDownScaler(&mockStore{}, &mockCluster{})
)

func TestDownScaler_Run(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Run() - must not panic")
		}
	}()

	ds.Run()
}

func TestDownScaler_mustScaleDown(t *testing.T) {
	testCases := []struct {
		deployName string
		want       bool
	}{
		{"error", false},
		{"timeout", true},
		{"notimeout", false},
	}

	for _, tc := range testCases {
		got := ds.mustScaleDown(tc.deployName, "")

		if got != tc.want {
			t.Errorf("mustScaleDown(%s, '') = %t; want = %t", tc.deployName, got, tc.want)
		}
	}
}
