# How works proxless

## Definitions

- `service` == `kubernetes service`
- `deployment` == `kubernetes deployment`

## In 1 minute

The main idea when building proxless is to be as simple as possible.  
You don't wanna install a Custom Resource Definition, you don't want to have a big stack, you just want something to scale up and down your deployment based on-demand.  
Proxless is a simple proxy written in golang and does not consume much resources. You don't need anything else.

You just need to add an annotation on the services you want to make on-demand and Proxless will handle the rest.

Start now and deploy the [helm chart](../helm/README.md).  

### Examples

Check the [example](../examples). Replace `hello-world.proxless.kintohub.net` with your own domain for testing.  
_Notes: don't forget to configure your DNS_

## Namespace scoped or cluster wide
 
- **Namespace scoped**
    - env var `NAMESPACE` must not be empty - proxless will only look for services within this namespace.
    - a `Role` is required.  See [here](../helm/templates/role.yaml).
- **Cluster wide**
    - env var `NAMESPACE` is empty - proxless will look for any services in the cluster.
    - a `ClusterRole` is required. See [here](../helm/templates/clusterrole.yaml).

## Core Concepts

Proxless works with 4 core concepts:

- the **store** is an in-memory storage containing the information for each route.
- the **proxy**, responsible for forwarding the requests to the correct back end and scaling up the deployment if needed.
- the **services engine**, responsible for retrieving the configuration from the services and configuring the deployments.
- the **downscaler**, responsible for downscaling the deployments when they are not used.

### The store

The store is an in-memory map and contain the information for each route.  
It will contain

- the namespace name where the deployment and service are
- the service name proxless will forward the request to
- the deployment name proxless will scale up and down
- the domain names / urls proxless proxless will proxy
- a timestamp of the last time the service has been requested

The logic of the store is available in [internal/store/inmemory/inmemory.go](../internal/store/inmemory/inmemory.go).

### The Proxy

The Proxy is the main process of proxless. It is the one responsible for the liveness of proxless. If the proxy crash, proxless restart.  
It is responsible for forwarding the requests to the correct back end and scaling up the deployment if needed.

Upon receiving a request to a specific URL, the proxy will

- retrieve the route information from the store
    - if the route is not in the store, it will return a `404`
- forward the request (with the headers) to the service
    - if the call fail (`could not resolve host` error), it will immediate try to scale down the deployment
    - when the deployment is ready, it will forward the request to the service
    - it will also the `lastUsed` timestamp of the route in the store
- if the above fail, it will return a `500`

The logic of the proxy is available in [internal/server/http/http.go](../internal/server/http/http.go).

### The Services Engine

The services engine run as a routine.  
It is responsible for looking at all the services, add them in the store and configure the deployments.

Upon creating/modifying a service, the services engine will

- check if the service contains the correct annotations
    - `proxless/domains`
        - external domain names associated to the service (separated with `,`)
        - example: `proxless/domains=example.io,www.example.io` 
    - `proxless/deployment`
        - deployment name used for scaling up and down the app
        - example: `hello-world`
    - (check the [example](../examples))
- if the service is compatible
    - it will add all the information into the store (see the [store section](#the-store))
    - it will add the label `proxless=true` to the deployment (for the [downscaler](#the-downscaler))
        - if the deployment does not exist, it will still add the information to the store so that the forwarding still works
        - it resyncs all the services every 30 seconds so the deployment can be picked up later

Upon deleting a service, the service engine will delete all its route information from the store and remove the proxless label from its deployment.

The logic of the services engine is available in [internal/cluster/kube/servicesinformer.go](../internal/cluster/kube/servicesinformer.go).

### The DownScaler

The DownScaler run as a routine.  
It is responsible for downscaling the deployment when the service has not been called for a while (configurable).

Every `N` seconds (configurable), the downscaler will

- retrieve all the deployments with the label `proxless=true`
- loop through each deployment that are running
    - retrieve its route information from the store
    - check if its `lastUsed` timestamp is > `timeout` (configurable)
        - if yes, it will scale down the deployment

The logic of the downscaler is available in the `RunDownScaler` func from [internal/controller/controller.go](../internal/controller/controller.go).

## TODO

### Scalability and AutoSync

Each replica of proxless run the same code so they should normally have the same store at any time.  
However, the `lastUsed` timestamp is only updated on the replica that receive the request. Therefore, we need a way to synchronise all the replicas with the latest `lastUsed`.  
We will persist it as a deployment annotation: `proxless/last-used=[timestamp]`.  

AutoSync will run as a routine is will be responsible for the above. 

Every `N` seconds (configurable), Autosync will

- loop through all the routes in the store
- retrieve the associated deployment
    - if the deployment does not have the annotation, it will add it
    - if the deployment have the annotation
        - if `lastUsed` annotation is newer than `lastUsed` store, it will update the store
        - else, it will update the deployment annotation

/!\ Update in kubernetes are not isolated.    
Indeed, if `replica 0` and `replica 1` retrieve the information at the same time, they will "conflict" each other by persisting their own `lastUsed` into the deployment.  
Therefore `N` has to be chosen carefully and much lower (half?) than the downscaler interval.