package logger

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
)

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
		_ = os.Setenv("LOG_LEVEL", tc.logLevel)

		InitLogger()

		if zerolog.GlobalLevel() != tc.want {
			t.Errorf("InitLogger(); GlobalLevel = %s; want = %s", zerolog.GlobalLevel(), tc.want)
		}
	}
}
