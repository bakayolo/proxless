package config

import (
	"os"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	// make sure it does not panic
	LoadEnvVars()
}

func Test_parseString(t *testing.T) {
	env := "env"
	testCases := []struct {
		value, defaultValue, want string
		mustPanic                 bool
	}{
		{"", "", "", true},
		{"something", "somethingelse", "something", false},
		{"", "somethingelse", "somethingelse", false},
	}

	for _, tc := range testCases {
		_ = os.Setenv(env, tc.value)
		got := assertParseStringPanic(t, env, tc.defaultValue, tc.mustPanic)

		if got != tc.want {
			t.Errorf("parseString(%s, %s) = %s; want = %s",
				tc.value, tc.defaultValue, got, tc.want)
		}
	}
}

func assertParseStringPanic(t *testing.T, key, defaultValue string, mustPanic bool) string {
	defer func() {
		if r := recover(); (r != nil) != mustPanic {
			t.Errorf("parseString(%s, %s); panic = %t; mustPanic = %t",
				key, defaultValue, r != nil, mustPanic)
		}
	}()

	return parseString(key, defaultValue)
}

func Test_parseInt(t *testing.T) {
	env := "env"
	testCases := []struct {
		value, defaultValue string
		want                int
		mustPanic           bool
	}{
		{"", "", 0, true},
		{"", "somethingelse", 0, true},
		{"something", "somethingelse", 0, true},
		{"1", "2", 1, false},
		{"", "2", 2, false},
	}

	for _, tc := range testCases {
		_ = os.Setenv(env, tc.value)
		got := assertParseIntPanic(t, env, tc.defaultValue, tc.mustPanic)

		if got != tc.want {
			t.Errorf("parseInt(%s, %s) = %d; want = %d",
				tc.value, tc.defaultValue, got, tc.want)
		}
	}
}

func assertParseIntPanic(t *testing.T, key, defaultValue string, mustPanic bool) int {
	defer func() {
		if r := recover(); (r != nil) != mustPanic {
			t.Errorf("parseInt(%s, %s); panic = %t; mustPanic = %t",
				key, defaultValue, r != nil, mustPanic)
		}
	}()

	return parseInt(key, defaultValue)
}
