package server

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes/upscaler"
	"kube-proxless/internal/store/inmemory"
	"net"
)

var httpClient = fasthttp.Client{
	MaxConnsPerHost: config.MaxConsPerHost,
}

func StartServer() {
	addr := fmt.Sprintf(":%s", config.Port)
	log.Info().Msgf("Proxless listening to %s", addr)

	server := fasthttp.Server{
		Name:    "proxless",
		Handler: requestHandler,
	}
	log.Fatal().Err(server.ListenAndServe(addr))
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.Header = ctx.Request.Header
	req.SetBody(ctx.Request.Body())

	// TODO do it globally
	store := inmemory.NewInMemoryStore()

	host := parseHost(ctx)
	route, err := store.GetRouteByDomain(host)
	if err != nil {
		log.Error().Err(err).Msgf("Could not find domain '%s' with parsed url '%s' in the store", ctx.Host(), host)
		ctx.Response.SetStatusCode(404)
		ctx.Response.SetBodyString(fmt.Sprintf("Domain %s not found", ctx.Host()))
	} else { // the route exists so we should have a deployment attached to the service
		service := route.GetService()
		namespace := route.GetNamespace()
		port := route.GetPort()

		origin := fmt.Sprintf("%s.%s:%s", service, namespace, port)
		req.SetHost(origin)

		if err := httpClient.Do(req, res); err != nil { // First try
			log.Debug().Msg("Error forwarding the request - Try scaling up the deployment")

			// the deployment is scaled down, let's scale it up
			deployment := route.GetDeployment()
			if err := upscaler.ScaleUpDeployment(deployment, namespace); err != nil {
				forwardError(ctx, err)
			} else { // Second try with the deployment scaled up
				if err := httpClient.Do(req, res); err != nil {
					forwardError(ctx, err)
				} else {
					forwardRequest(ctx, res)
				}
			}
		} else {
			forwardRequest(ctx, res)
		}

		_ = store.UpdateLastUse(host)
	}
}

func forwardRequest(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	log.Debug().Msg("Request forwarded")
	ctx.Response.SetBodyString(string(res.Body()))
	ctx.Response.Header = res.Header
}

func forwardError(ctx *fasthttp.RequestCtx, err error) {
	log.Error().Err(err).Msg("Error forwarding the request")
	ctx.Response.SetBodyString("Error in the server")
	ctx.Response.SetStatusCode(500)
}

func parseHost(ctx *fasthttp.RequestCtx) string {
	host, _, err := net.SplitHostPort(string(ctx.Host()))
	if err != nil { // no port
		return string(ctx.Host())
	}
	return host
}
