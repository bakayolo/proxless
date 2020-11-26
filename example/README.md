# Proxless example

## Requirement

Follow the [main README](../README.md) and have proxless deployed either globally or scoped in a namespace.

## Example

The example is a NGINX server pod doing a `proxy_pass` to a hello-world microservice.

### NGINX service

The NGINX service annotations are 

```yaml
proxless/domains: "example.io,www.example.io"
proxless/deployment: "frontend"
proxless/ttl-seconds: "120"
proxless/readiness-timeout-seconds: "30"
```

So the NGINX service will be accessible through proxless using example.io and www.example.io and it will scale up and down the deployment `frontend`.  
Additionally, it will be accessible internally through `frontend-proxless.[YOUR NAMESPACE]` and `frontend-proxless.[YOUR NAMESPACE].svc.cluster.local`.

This service will not use the default configuration (environment variable) for Time To Live and Readiness Timeout.  
The NGINX deployment will be scaled down after 120 seconds being not used and will timeout after 30 seconds when scaling up.

### Hello World service

The Hello World annotation is 

```yaml
proxless/deployment: "hello"
```

The hello world service does not contain any domain so will only be accessible through `hello-proxless.[YOUR NAMESPACE]` and `hello-proxless.[YOUR NAMESPACE].svc.cluster.local`.

### Scoped Namespace

If proxless is scoped to a namespace, both service will also be accessible internally through `frontend-proxless` for NGINX and `hello-proxless` for the hello world microservice.

## Deploy the example

### Kubernetes

```shell script
$ kubectl apply -f kubectl/example.yaml
```

## Port-Forward

Port-forward to your proxless deployment.

### Kubernetes

```shell script
$ kubectl port-forward svc/proxless 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

## Call it

### Commons

```shell script
$ export YOUR_NAMESPACE="proxless"

$ curl -H "Host: www.example.io" localhost:8080
{"message":"Hello"}

$ curl -H "Host: example.io" localhost:8080
{"message":"Hello"}

$ curl -H "Host: hello-proxless.${YOUR_NAMESPACE}.svc.cluster.local" localhost:8080
{"message":"Hello"}

$ curl -H "Host: frontend-proxless.${YOUR_NAMESPACE}" localhost:8080
{"message":"Hello"}
```

### Namespace scoped

```shell script
$ curl -H "Host: frontend-proxless" localhost:8080
{"message":"Hello"}

$ curl -H "Host: hello-proxless" localhost:8080
{"message":"Hello"}
```