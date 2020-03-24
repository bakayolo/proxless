# Proxless

> Reduce your kubernetes cost by making all your deployments serverless with Proxless

Proxless is an open-source proxy made for Kubernetes.
Every deployment/service that are fronted by Proxless will be made serverless.
Proxless scale to 0 your deployments after N seconds idle and scale them back to 1 after a call to the URL.

## Installation

A helm chart will be available soon.

## Development Setup

Duplicate the `.env.example` file into a `.env` file and modify the variables accordingly

```shell script
KUBE_CONFIG_PATH= ## Path your your `kube.config` file
LOG_LEVEL=DEBUG
PORT=8080
MAX_CONS_PER_HOST=10000 ## Max number of concurrent connections that can be forwarded to the origin servers

NAMESPACE= ## Kubernetes namespace
## If specified, proxless will watch one namespace. A Kubernetes Role is needed.
## If empty, proxless will watch all the namespaces. A Kubernetes ClusterRole is needed.

SERVERLESS_TTL_SECONDS=10 ## Time to leave in seconds for your serverless deployments

## When Proxless is scaling up a deployment
READINESS_POLL_TIMEOUT_SECONDS=30 ## Proxless wait this time before timing out
READINESS_POLL_INTERVAL_SECONDS=1 ## Proxless check interval
```

Then run

```shell script
$ go run cmd/main.go
```

## Meta

Benjamin APPREDERISSE - [@benhazard42](https://twitter.com/benhazard42)

Distributed under the MIT license. See ``LICENSE`` for more information.

## Contributing

1. Fork it (<https://github.com/bappr/kube-proxless/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request