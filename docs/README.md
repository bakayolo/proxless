# How works proxless

## Definitions

- `service` == `kubernetes service`
- `deployment` == `kubernetes deployment`

## Core Concepts

Proxless works with 4 core concepts:

- the **memory** store is an in-memory map containing the information for each route.
- the **proxy**, responsible for forwarding the requests to the correct back end and scaling up the deployment if needed.
- the **services engine**, responsible for retrieving the configuration from the services and configuring the deployments.
- the **downscaler**, responsible for downscaling the deployments when they are not used.
- the **pubsub** system is used to update the latest used time for each request on each proxless replicas.

### Memory

The in-memory map contains the information for each route.  
It will contain

- the namespace name where the deployment and service are
- the service name proxless will forward the request to
- the deployment name proxless will scale up and down
- the domain names / urls proxless proxless will proxy
- a timestamp of the last time the service has been requested

The logic of the in-memory map is available in [internal/memory/memory.go](../internal/memory/memory.go).

### The Proxy

The Proxy is the main process of proxless. It is the one responsible for the liveness of proxless. If the proxy crash, proxless restart.  
It is responsible for forwarding the requests to the correct back end and scaling up the deployment if needed.

Upon receiving a request to a specific URL, the proxy will

- retrieve the route information from the memory
    - if the route is not in memory, it will return a `404`
- forward the request (with the headers) to the service
    - if the call fail (`could not resolve host` error), it will immediate try to scale down the deployment
    - when the deployment is ready, it will forward the request to the service
    - it will also update the `lastUsed` timestamp of the route in the memory
- if the above fail, it will return a `500`

The logic of the proxy is available in [internal/server/http/http.go](../internal/server/http/http.go).

### The Services Engine

The services engine run as a routine.  
It is responsible for looking at all the services, add them in memory and configure the deployments.

Upon creating/modifying a service, the services engine will

- check if the service contains the required annotation
    - `proxless/deployment`
        - deployment name used for scaling up and down the app
        - example: `hello`
- check if the service contains additional annotation for domains
    - `proxless/domains`
        - external domain names associated to the service (separated with `,`)
        - example: `proxless/domains=example.io,www.example.io` 
- if the service is compatible
    - it will add all the information into the memory (see the [memory section](#memory))
    - it will add the label `proxless=true` to the deployment (for the [downscaler](#the-downscaler))
        - if the deployment does not exist, it will still add the information into memory so that the forwarding still works
        - it resyncs all the services every 30 seconds so the deployment can be picked up later
    - it will create new service `[SERVICE]-proxless` that you can use to access your service internally through proxless.

Upon deleting a service, the service engine will delete all its route information from the memory, remove the proxless label from its deployment and remove the proxless service.

The logic of the services engine is available in [internal/cluster/kube/servicesinformer.go](../internal/cluster/kube/servicesinformer.go).

### The DownScaler

The DownScaler run as a routine.  
It is responsible for downscaling the deployment when the service has not been called for a while (configurable).

Every `N` seconds (configurable), the downscaler will

- retrieve all the deployments with the label `proxless=true`
- loop through each deployment that are running
    - retrieve its route information from the memory
    - check if its `lastUsed` timestamp is > `timeout` (configurable)
        - if yes, it will scale down the deployment

The logic of the downscaler is available in the `RunDownScaler` func from [internal/controller/controller.go](../internal/controller/controller.go).

### PubSub

The pubsub system is used to update the latest used time for each request on each proxless replicas.  
It is optional and currently use Redis. Without the pubsub, proxless is not fully HA.  

- Upon receiving a new proxless compatible service (the services engine), every replica subscribe to a channel corresponding to the service id in the pubsub system.  
- When a service is being called (the proxy), proxless will `PUBLISH` the `lastUsed` time attached to the service id to the pubsub system.
- Upon receiving a message from the pub/sub system, the replica will update the memory store.

This guarantee an eventual consistency by making sure that every replica connected to the pubsub system will always end up with the latest `lastUsed` time for each request.

The logic of the pubsub is available in [internal/pubsub/redis/redis.go](../internal/pubsub/redis/redis.go).