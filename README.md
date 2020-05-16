# Proxless

> Reduce your kubernetes cost by making all your deployments **on-demand** with Proxless
> Deploy Proxless in front of your services and it will scale down the associated deployments when they are not requested and scale them back up when they are.

**No need a CRD, no need a huge stack, the proxless deployment are the only things u need.**

## Disclaimer

Proxless is provided in alpha mode.  
Using it on your production cluster is done at your own risks.

## In 1 minute

Proxless is a simple proxy written in golang and consume a minimum of resources.  
You don't need to run anything other than proxless deployment and it will not modify your existing resources.

Proxless looks for the services in the cluster that have a specific annotation and scale up and down their associated deployment. 

_Note: in order for proxless to be fully high available, all the replicas need to sync up the `lastUsed` time for each request between each other.
In order to achieve that, a non persistent standalone redis is needed. This configuration is fully optional and provided in the helm chart._

Check the [documentation](docs) for more information.

## Namespace scoped or cluster wide
 
- **Namespace scoped**
    - env var `NAMESPACE_SCOPED` must be `true` - proxless will only look for services within its namespace.
    - a `Role` is required.
- **Cluster wide**
    - env var `NAMESPACE_SCOPED` is `false - proxless will look for any services in the cluster.
    - a `ClusterRole` is required. See [here](deploy/helm/templates/clusterrole.yaml).

## Quickstart

### Kubectl

```shell script
$ kubectl apply -f deploy/kubectl/proxless.yaml
```

This will deploy a proxless scoped to your namespace.  
Use the helm chart below to make it cluster wide.

### Helm

You can use our [helm chart](deploy/helm/README.md) for a more configurable approach.  
With the chart, you will be able to configure the High Availability and choose if you want proxless to be installed cluster wide or not.

## Test it

Deploy the [example](example/kubectl/example.yaml).  
It's a basic nginx pod doing a `proxy_pass` to a hello-world api pod.  
By default, the 2 pods are scaled down.  

```shell script
$ kubectl apply -f example/kubectl/example.yaml
```

Port-forward to your proxless deployment.

```shell script
$ kubectl port-forward svc/proxless 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

Call it

```shell script
$ curl -H "Host: www.example.io" localhost:8080
{"message":"Hello"}

$ curl -H "Host: example.io" localhost:8080
{"message":"Hello"}
```

More information [here](example/kubectl/README.md)

## Development Setup

Duplicate the `.env.example` file into a `.env` file and modify the variables accordingly

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
