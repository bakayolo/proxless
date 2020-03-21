package server

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/commons"
	"kube-proxless/internal/config"
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

	req.Header = ctx.Request.Header
	req.SetBody(ctx.Request.Body())

	host := parseHost(ctx)
	if host == "" {
		ctx.Response.SetStatusCode(404)
		ctx.Response.SetBodyString(fmt.Sprintf("Domain %s not found", ctx.Host()))
	} else {
		req.SetHost(host)

		if err := httpClient.Do(req, res); err != nil {
			log.Error().Err(err).Msg("Error forwarding the request")
		} else {
			log.Debug().Msg("Request forwarded")

			ctx.Response.SetBodyString(string(res.Body()))
			ctx.Response.Header = res.Header
		}
	}
}

func parseHost(ctx *fasthttp.RequestCtx) string {
	u, err := url.Parse(string(ctx.Host()))
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing URL %s", ctx.Host())
		return ""
	}
	return commons.GetRoute(u.Scheme)
}
