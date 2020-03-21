package server

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"kube-proxless/internal/config"
)
import "github.com/rs/zerolog/log"

var httpClient = fasthttp.Client{
	MaxConnsPerHost: config.MaxConsPerHost,
}

func StartServer() {
	addr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Proxless listening to %s", addr)

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
	req.SetRequestURI("https://www.google.com")

	if err := httpClient.Do(req, res); err != nil {
		log.Err(err).Msg("Error forwarding the request")
	} else {
		log.Debug().Msg("Request forwarded")

		ctx.Response.SetBodyString(string(res.Body()))
		ctx.Response.Header = res.Header
	}
}
