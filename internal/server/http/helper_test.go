package http

import (
	"errors"
	"github.com/valyala/fasthttp"
)

type mockFastHTTP struct {
	doMustFail bool
}

func (*mockFastHTTP) listenAndServe(host string, requestHandler func(ctx *fasthttp.RequestCtx)) {}

func (m *mockFastHTTP) do(req *fasthttp.Request, resp *fasthttp.Response) error {
	if m.doMustFail {
		return errors.New("do must fail")
	}

	return nil
}
