package server

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/config"
	"kube-proxless/internal/kubernetes"
	"kube-proxless/internal/store"
	"net/url"
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

	host := parseHost(ctx)
	if host == "" {
		ctx.Response.SetStatusCode(404)
		ctx.Response.SetBodyString(fmt.Sprintf("Domain %s not found", ctx.Host()))
	} else { // the route exists so we should have a deployment attached to the service
		origin := store.GetRouteOrigin(host)
		req.SetHost(origin)

		// First try
		if err := httpClient.Do(req, res); err != nil {
			log.Debug().Msg("Error forwarding the request - Scaling up the deployment")
			// Maybe the deployment is scaled down, let's scale it up
			if err := kubernetes.ScaleUp(store.GetRouteLabel(host), store.GetRouteNamespace(host)); err != nil {
				errorForwarded(ctx, err)
			} else { // Second try with the deployment scaled up
				if err := httpClient.Do(req, res); err != nil {
					errorForwarded(ctx, err)
				} else {
					requestForwarded(ctx, res)
				}
			}
		} else {
			requestForwarded(ctx, res)
		}
	}
}

func requestForwarded(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	log.Debug().Msg("Request forwarded")
	ctx.Response.SetBodyString(string(res.Body()))
	ctx.Response.Header = res.Header
}

func errorForwarded(ctx *fasthttp.RequestCtx, err error) {
	log.Error().Err(err).Msg("Error forwarding the request")
	ctx.Response.SetBodyString("Error in the server")
	ctx.Response.SetStatusCode(500)
}

func parseHost(ctx *fasthttp.RequestCtx) string {
	u, err := url.Parse(string(ctx.Host()))
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing URL %s", ctx.Host())
		return ""
	}
	//TODO why do I need to use `u.Scheme` instead of `u.Host`?
	return u.Scheme
}
