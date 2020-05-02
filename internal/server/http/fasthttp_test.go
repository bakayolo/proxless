package http

import (
	"testing"
)

func Test_newFastHTTP(t *testing.T) {
	want := 1

	fastHTTP := newFastHTTP(want)

	if fastHTTP.client.MaxConnsPerHost != want {
		t.Errorf("newFastHTTP(%d); maxConnsPerHost == %d; want %d",
			want, fastHTTP.client.MaxConnsPerHost, want)
	}
}
