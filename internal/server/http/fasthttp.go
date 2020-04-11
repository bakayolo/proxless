package http

import (
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/logger"
)

type fastHTTP struct {
	client fasthttp.Client
}

type fastHTTPInterface interface {
	listenAndServe(host string, requestHandler func(ctx *fasthttp.RequestCtx))
	do(req *fasthttp.Request, resp *fasthttp.Response) error
}

func newFastHTTP(maxConsPerHost int) *fastHTTP {
	return &fastHTTP{
		client: fasthttp.Client{
			MaxConnsPerHost: maxConsPerHost,
		},
	}
}

func (*fastHTTP) listenAndServe(host string, requestHandler func(ctx *fasthttp.RequestCtx)) {
	server := fasthttp.Server{
		Name:    "proxless-http",
		Handler: requestHandler,
	}
	logger.Fatalf(server.ListenAndServe(host), "Error starting the server")
}

func (f *fastHTTP) do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return f.client.Do(req, resp)
}
