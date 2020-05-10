package config

import (
	"os"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	// make sure it does not panic
	LoadEnvVars()
}

func Test_getString(t *testing.T) {
	env := "env"
	testCases := []struct {
		value, defaultValue, want string
	}{
		{"", "", ""},
		{"something", "somethingelse", "something"},
		{"", "somethingelse", "somethingelse"},
	}

	for _, tc := range testCases {
		_ = os.Setenv(env, tc.value)
		got := getString(env, tc.defaultValue)

		if got != tc.want {
			t.Errorf("getString(%s, %s) = %s; want = %s", tc.value, tc.defaultValue, got, tc.want)
		}
	}
}

func Test_getInt(t *testing.T) {
	env := "env"
	testCases := []struct {
		value              string
		defaultValue, want int
		mustPanic          bool
	}{
		{"", 0, 0, false},
		{"", 1, 1, false},
		{"something", 2, 2, true},
		{"1", 2, 1, false},
	}

	for _, tc := range testCases {
		_ = os.Setenv(env, tc.value)
		got := assertParseIntPanic(t, env, tc.defaultValue, tc.mustPanic)

		if !tc.mustPanic && got != tc.want {
			t.Errorf("getInt(%s, %d) = %d; want = %d", env, tc.defaultValue, got, tc.want)
		}
	}
}

func assertParseIntPanic(t *testing.T, key string, defaultValue int, mustPanic bool) int {
	defer func() {
		if r := recover(); (r != nil) != mustPanic {
			t.Errorf("getInt(%s, %d); panic = %t; mustPanic = %t",
				key, defaultValue, r != nil, mustPanic)
		}
	}()

	return getInt(key, defaultValue)
}

func Test_getBool(t *testing.T) {
	env := "env"
	testCases := []struct {
		value              string
		defaultValue, want bool
		mustPanic          bool
	}{
		{"", true, true, false},
		{"false", true, false, false},
		{"something", true, true, true},
	}

	for _, tc := range testCases {
		_ = os.Setenv(env, tc.value)
		got := assertParseBoolPanic(t, env, tc.defaultValue, tc.mustPanic)

		if !tc.mustPanic && got != tc.want {
			t.Errorf("getBool(%s, %t) = %t; want = %t", env, tc.defaultValue, got, tc.want)
		}
	}
}

func assertParseBoolPanic(t *testing.T, key string, defaultValue, mustPanic bool) bool {
	defer func() {
		if r := recover(); (r != nil) != mustPanic {
			t.Errorf("getBool(%s, %t); panic = %t; mustPanic = %t",
				key, defaultValue, r != nil, mustPanic)
		}
	}()

	return getBool(key, defaultValue)
}
