package http

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/cluster/fake"
	"kube-proxless/internal/config"
	"kube-proxless/internal/controller"
	"kube-proxless/internal/memory"
	"kube-proxless/internal/model"
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
	server := NewHTTPServer(controller.NewController(memory.NewMemoryMap(), fake.NewCluster(), nil))
	server.client = &mockFastHTTP{}

	// make sure it does not panic
	server.Run()
}

func TestHTTPServer_requestHandler(t *testing.T) {
	mem := memory.NewMemoryMap()
	server := NewHTTPServer(controller.NewController(mem, fake.NewCluster(), nil))

	testCases := []struct {
		host       string
		doMustFail bool
		want       int
	}{
		{"", false, 404},
		{"mock.io", false, 200},
		{"mock.io", true, 500},
	}

	route, err := model.NewRoute(
		"mock-id", "mock-svc", "", "mock-deploy", "mock-ns", []string{"mock.io"}, nil, nil)
	assert.NoError(t, err)
	route.SetIsRunning(false)

	// add route in the memory
	err = mem.UpsertMemoryMap(route)
	assert.NoError(t, err)

	for _, tc := range testCases {
		req := fasthttp.AcquireRequest()
		req.SetHost(tc.host)

		ctx := &fasthttp.RequestCtx{Request: *req}

		server.client = &mockFastHTTP{doMustFail: tc.doMustFail}
		server.requestHandler(ctx)

		assert.Equal(t, ctx.Response.StatusCode(), tc.want, fmt.Sprintf("requestHandler(%s);", tc.host))

		fasthttp.ReleaseRequest(req)
	}
}

func TestHTTPServer_forward404Error(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Response: fasthttp.Response{},
	}
	forward404Error(ctx, nil, "test")

	want := 404

	assert.Equal(t, ctx.Response.StatusCode(), want)
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

	assert.Equal(t, ctx.Response.StatusCode(), statusCodeWant)

	assert.Equal(t, string(ctx.Response.Body()), bodyWant)
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
