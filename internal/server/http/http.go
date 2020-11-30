package http

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"k8s.io/apimachinery/pkg/util/wait"
	"kube-proxless/internal/config"
	"kube-proxless/internal/controller"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/server/utils"
	"time"
)

type httpServer struct {
	controller controller.Interface
	client     fastHTTPInterface
	host       string
}

func NewHTTPServer(controller controller.Interface) *httpServer {
	return &httpServer{
		controller: controller,
		client:     newFastHTTP(config.MaxConsPerHost),
		host:       fmt.Sprintf(":%s", config.Port),
	}
}

func (s *httpServer) Run() {
	logger.Infof("Proxless listening to %s", s.host)

	s.client.listenAndServe(s.host, s.requestHandler)
}

func (s *httpServer) requestHandler(ctx *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.Header = ctx.Request.Header
	req.SetBody(ctx.Request.Body())

	logger.Debugf("Received request %s", ctx.Host())

	host := utils.ParseHost(string(ctx.Host()))
	route, err := s.controller.GetRouteByDomainFromMemory(host)
	if err != nil {
		forward404Error(ctx, err, host)
	} else { // the route exists so we should have a deployment attached to the service
		service := route.GetService()
		namespace := route.GetNamespace()
		port := route.GetPort()

		origin := fmt.Sprintf("%s.%s:%s", service, namespace, port)
		req.SetHost(origin)

		// update before because it's gonna take some time to scale up the deployment
		_ = s.controller.UpdateLastUsedInMemory(route.GetId())

		err := s.client.do(req, res)

		if err != nil { // First try, the deployment might be scaled down
			readinessTimeoutSeconds := config.DeploymentReadinessTimeoutSeconds
			if route.GetReadinessTimeoutSeconds() != nil {
				readinessTimeoutSeconds = *route.GetReadinessTimeoutSeconds()
			}

			if route.GetIsRunning() {
				logger.Debugf("Error forwarding the request %s - deployment is already running, we just wait", ctx.Host())

				err := waitForResponse(s, req, res, readinessTimeoutSeconds)

				if err != nil {
					forwardError(ctx, err)
				}
			} else {
				logger.Debugf("Error forwarding the request %s - Try scaling up the deployment", ctx.Host())
				// we update the isRunning cuz to make sure we don't have multiple tentatives of waking up the deployment
				// at the same time - otherwise that would overload the kubernetes api
				_ = s.controller.UpdateIsRunningInMemory(route.GetId())

				err := s.controller.ScaleUpDeployment(route.GetDeployment(), namespace, readinessTimeoutSeconds)

				if err != nil {
					forwardError(ctx, err)
				} else { // Second try with the deployment scaled up
					err := s.client.do(req, res)

					if err != nil {
						forwardError(ctx, err)
					} else {
						forwardRequest(ctx, res)
					}
				}
			}
		} else {
			forwardRequest(ctx, res)
		}

		// update after because it took some time to scale up the deployment
		// TODO see this, I don't like updating it twice
		_ = s.controller.UpdateLastUsedInMemory(route.GetId())
	}
}

// we call the backend regularly to see if the app is responding or not
// TODO implement some sort of queuing system to make sure the request are being sent in order
func waitForResponse(s *httpServer, req *fasthttp.Request, res *fasthttp.Response, readinessTimeoutSeconds int) error {
	err := wait.PollImmediate(1*time.Second, time.Duration(readinessTimeoutSeconds)*time.Second, func() (bool, error) {
		err := s.client.do(req, res)

		if err == nil {
			return true, nil
		} else {
			return false, nil
		}
	})

	return err
}

func forward404Error(ctx *fasthttp.RequestCtx, err error, host string) {
	logger.Errorf(err, "Could not find domain '%s' with parsed url '%s' in memory", ctx.Host(), host)
	ctx.Response.SetStatusCode(404)
	ctx.Response.SetBodyString(fmt.Sprintf("Domain %s not found", ctx.Host()))
}

func forwardRequest(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	logger.Debugf("Request %s forwarded", ctx.Host())
	ctx.Response.SetBodyString(string(res.Body()))
	ctx.Response.Header = res.Header
	ctx.Response.SetStatusCode(res.StatusCode())
}

func forwardError(ctx *fasthttp.RequestCtx, err error) {
	logger.Errorf(err, "Error forwarding %s request", ctx.Host())
	ctx.Response.SetBodyString("Error in the server")
	ctx.Response.SetStatusCode(500)
}
