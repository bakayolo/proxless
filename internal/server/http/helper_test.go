package http

import (
	"errors"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/model"
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

type mockController struct{}

func (*mockController) GetRouteByDomainFromStore(domain string) (*model.Route, error) {
	if domain == "" {
		return nil, errors.New("route not found")
	}

	deploy := "mock"
	if domain != "mock" {
		deploy = "err"
	}

	return model.NewRoute("mock", "mock", "", deploy, "mock", []string{"mock.io"})
}

func (*mockController) UpdateLastUseInStore(domain string) error {
	return nil
}

func (*mockController) ScaleUpDeployment(name, namespace string) error {
	if name != "mock" {
		return errors.New("route not found")
	}

	return nil
}
