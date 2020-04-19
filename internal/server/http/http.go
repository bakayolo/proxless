package http

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/config"
	"kube-proxless/internal/controller"
	"kube-proxless/internal/server/utils"
)

type HTTPServer struct {
	controller controller.ControllerInterface
	client     fastHTTPInterface
	host       string
}

func NewHTTPServer(controller *controller.Controller) *HTTPServer {
	return &HTTPServer{
		controller: controller,
		client:     newFastHTTP(config.MaxConsPerHost),
		host:       fmt.Sprintf(":%s", config.Port),
	}
}

func (s *HTTPServer) Run() {
	log.Info().Msgf("Proxless listening to %s", s.host)

	s.client.listenAndServe(s.host, s.requestHandler)
}

func (s *HTTPServer) requestHandler(ctx *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.Header = ctx.Request.Header
	req.SetBody(ctx.Request.Body())

	host := utils.ParseHost(string(ctx.Host()))
	route, err := s.controller.GetRouteByDomainFromStore(host)
	if err != nil {
		s.forward404Error(ctx, err, host)
	} else { // the route exists so we should have a deployment attached to the service
		service := route.GetService()
		namespace := route.GetNamespace()
		port := route.GetPort()

		origin := fmt.Sprintf("%s.%s:%s", service, namespace, port)
		req.SetHost(origin)

		if err := s.client.do(req, res); err != nil { // First try
			log.Debug().Msg("Error forwarding the request - Try scaling up the deployment")

			// the deployment is scaled down, let's scale it up
			deployment := route.GetDeployment()
			if err := s.controller.ScaleUpDeployment(deployment, namespace); err != nil {
				s.forwardError(ctx, err)
			} else { // Second try with the deployment scaled up
				if err := s.client.do(req, res); err != nil {
					s.forwardError(ctx, err)
				} else {
					s.forwardRequest(ctx, res)
				}
			}
		} else {
			s.forwardRequest(ctx, res)
		}

		_ = s.controller.UpdateLastUseInStore(host)
	}
}

func (s *HTTPServer) forward404Error(ctx *fasthttp.RequestCtx, err error, host string) {
	log.Error().Err(err).Msgf("Could not find domain '%s' with parsed url '%s' in the store", ctx.Host(), host)
	ctx.Response.SetStatusCode(404)
	ctx.Response.SetBodyString(fmt.Sprintf("Domain %s not found", ctx.Host()))
}

func (s *HTTPServer) forwardRequest(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	log.Debug().Msg("Request forwarded")
	ctx.Response.SetBodyString(string(res.Body()))
	ctx.Response.Header = res.Header
}

func (s *HTTPServer) forwardError(ctx *fasthttp.RequestCtx, err error) {
	log.Error().Err(err).Msg("Error forwarding the request")
	ctx.Response.SetBodyString("Error in the server")
	ctx.Response.SetStatusCode(500)
}
