package utils

import "testing"

func TestParseHost(t *testing.T) {
	testCases := []struct {
		host, want string
	}{
		{"example.com", "example.com"},
		{"example.com:80", "example.com"},
	}

	for _, tc := range testCases {
		got := ParseHost(tc.host)

		if got != tc.want {
			t.Errorf("ParseHost(%s) = %s; want %s", tc.host, got, tc.want)
		}
	}
}
