package http

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/cluster/fake"
	"kube-proxless/internal/config"
	"kube-proxless/internal/controller"
	"kube-proxless/internal/store/inmemory"
	"testing"
)

func TestNewHTTPServer(t *testing.T) {
	HTTPServer := NewHTTPServer(nil)

	if HTTPServer.host != fmt.Sprintf(":%s", config.Port) {
		t.Errorf("NewHTTPServer(nil); host == %s but must be %s",
			HTTPServer.host, fmt.Sprintf(":%s", config.Port))
	}
}

func TestHTTPServer_Run(t *testing.T) {
	server := NewHTTPServer(controller.NewController(inmemory.NewInMemoryStore(), fake.NewCluster()))
	server.client = &mockFastHTTP{}

	// make sure it does not panic
	server.Run()
}

func TestHTTPServer_requestHandler(t *testing.T) {
	store := inmemory.NewInMemoryStore()
	server := NewHTTPServer(controller.NewController(store, fake.NewCluster()))

	testCases := []struct {
		host       string
		doMustFail bool
		want       int
	}{
		{"", false, 404},
		{"mock.io", false, 200},
		{"mock.io", true, 500},
	}

	// add route in the store
	_ = store.UpsertStore(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns", []string{"mock.io"})

	for _, tc := range testCases {
		req := fasthttp.AcquireRequest()
		req.SetHost(tc.host)

		ctx := &fasthttp.RequestCtx{Request: *req}

		server.client = &mockFastHTTP{doMustFail: tc.doMustFail}
		server.requestHandler(ctx)

		if ctx.Response.StatusCode() != tc.want {
			t.Errorf("requestHandler(%s); statusCode = %d; want %d",
				tc.host, ctx.Response.StatusCode(), tc.want)
		}

		fasthttp.ReleaseRequest(req)
	}
}

func TestHTTPServer_forward404Error(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Response: fasthttp.Response{},
	}
	forward404Error(ctx, nil, "test")

	want := 404

	if got := ctx.Response.StatusCode(); got != want {
		t.Errorf("forward404Error(); status code == %d but must be %d", got, want)
	}
}

func TestHTTPServer_forwardRequest(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Response: fasthttp.Response{},
	}

	statusCodeWant := 200
	bodyWant := "testing 200"

	res := &fasthttp.Response{}
	res.SetStatusCode(statusCodeWant)
	res.SetBodyString(bodyWant)

	forwardRequest(ctx, res)

	if statusCodeGot := ctx.Response.StatusCode(); statusCodeGot != statusCodeWant {
		t.Errorf("forwardRequest(); status code == %d but must be %d", statusCodeGot, statusCodeWant)
	}

	if bodyGot := string(ctx.Response.Body()); bodyGot != bodyWant {
		t.Errorf("forwardRequest(); body == %s but must be %s", bodyGot, bodyWant)
	}
}

func TestHTTPServer_forwardError(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Response: fasthttp.Response{},
	}
	forwardError(ctx, nil)

	want := 500

	if got := ctx.Response.StatusCode(); got != want {
		t.Errorf("forwardError(); status code == %d but must be %d", got, want)
	}
}
