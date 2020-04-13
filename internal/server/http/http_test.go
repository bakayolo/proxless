package http

import (
	"github.com/valyala/fasthttp"
	"testing"
)

var server = HTTPServer{}

func TestHTTPServer_forward404Error(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Response: fasthttp.Response{},
	}
	server.forward404Error(ctx, nil, "test")

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

	server.forwardRequest(ctx, res)

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
	server.forwardError(ctx, nil)

	want := 500

	if got := ctx.Response.StatusCode(); got != want {
		t.Errorf("forwardError(); status code == %d but must be %d", got, want)
	}
}
