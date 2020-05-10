# Proxless

[proxless](https://github.com/bappr/kube-proxless) is an opensource proxy that use services annotations to scale up and down your deployments when they are not used.

To use it, add the `proxless/domains` and the `proxless/deployment` annotation to your Service resources.

## Introduction

This chart bootstraps a proxless deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.  
It also install a redis standalone for HA purpose.

## Prerequisites

Tested with
  - Kubernetes 1.14+

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release .
```

The command deploys proxless on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the proxless chart and their default values.

Parameter | Description | Default
--- | --- | ---
`image.repository` | container image repository | `bappr/proxless`
`image.tag` | container image tag | `latest`
`image.pullPolicy` | container image pull policy | `Always`
`logLevel` | proxless log level | `DEBUG`
`port` | port proxless is listening to | `8080`
`namespaceScoped` | is proxless working within a single namespace or across multiple namespaces | `true`
`env.MAX_CONS_PER_HOST` | max connections proxless can forward for a single host. More info [here](https://godoc.org/github.com/valyala/fasthttp#Client) | `10000`
`env.SERVERLESS_TTL_SECONDS` | time in seconds proxless waits before scaling down the app | `30`
`env.DEPLOYMENT_READINESS_TIMEOUT_SECONDS` | time in seconds proxless waits for the deployment to be ready when scaling up the app | false
`service.type` | kubernetes service type | `ClusterIP`
`ingress.enabled` | create a kubernetes ingress resource for calling proxless externally. | `false`
`ingress.annotations` | ingress annotations | `kubernetes.io/ingress.class: nginx`
`ingress.tls.enabled` | configure the ingress host to be https. | `false`
`ingress.tls.secret` | name of the secret containing the private key and certificate | `proxless-tls`
`ingress.host` | host to call proxless externally | `proxless.kintohub.net` 

These parameters can be passed via Helm's `--set` option
```console
$ helm install . --name my-release \
    --set service.type=LoadBalancer
```

_Notes: Enabling `ingress` is useless since it will proxy nowhere_

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install . --name my-release -f values.yaml
```