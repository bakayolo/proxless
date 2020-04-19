package config

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LoadEnvVars() - must not panic")
		}
	}()

	// only one test since all the environment variables have default values atm
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

func TestInitLogger(t *testing.T) {
	testCases := []struct {
		logLevel string
		want     zerolog.Level
	}{
		{"info", zerolog.InfoLevel},
		{"", zerolog.InfoLevel},
		{"error", zerolog.ErrorLevel},
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
	}

	for _, tc := range testCases {
		LogLevel = tc.logLevel

		got := InitLogger()
		if got != tc.want {
			t.Errorf("InitLogger() = %s; want = %s", got, tc.want)
		}
	}
}
