# How proxless works
  
> Proxless looks for any kubernetes services with specific annotations and make their associated deployment serverless.
 
It works with 2 modes:

- **Namespace scoped** - env var `NAMESPACE` is provided - proxless will only look for services within this namespace.
- **Global** - env var `NAMESPACE` is empty - proxless will look for any services in the cluster.


## ServicesEngine

The ServiceEngine is accessible [here](../internal/kubernetes/servicesengine).  
It run as a separate routine. It is responsible for looking at the services and storing the information in the store.  

### Get services

It will look for any kubernetes services with the 2 annotations

- `proxless/domains` - external domain names associated to the service (separated with `,`) - example: `proxless/domains=example.io,www.example.io` 
- `proxless/deployment` - deployment name used for scaling up and down the app - example: `hello-world`

### Label deployment

It will label the associated deployed with `proxless=true`.  
This will be used by the downscaler to get all the deployment managed by proxless.

### Add service information into the store

It will create an object [`Route`](../internal/store/routes.go) and add it to an in-memory store (map) using the UUID of the service.  

Example:
```shell script
uuid            -> Route Object containing the information of the service

domain 1        -> uuid
domain 2        -> uuid

deployment name -> uuid
```

## Proxy

The Proxy is the main process of proxless and is accessible [here](../internal/server/server.go).  
It is a [fasthttp](https://github.com/valyala/fasthttp) server and is responsible for routing the request, updating the `last used` timestamp and calling the upscaler if needed.

### Route request

Since the store contains the `domains`, it's easy to get the service information for a specific domain.

Example:
```gotemplate
if host == `example.io` {
  uuid := get_uid_with_domain(`example.io`)
  route := get_route_with_uid(uid)
  update_last_use(route)
  ok := proxy_to(route)
  if !ok {
    upscale(route)
    proxy_to(route)
  }
}
```

### Upscale

It will retrieve the deployment from kubernetes and update the `replicas` to `1`.

## Downscaler

The Dowscaler is accessible [here](../internal/kubernetes/downscaler).  
It run as a separate routine. It is responsible for downscaling the deployments that are considered idle (not used after N seconds).  
Since the store contains `deployment name`, it's easy to get the service information for a specific deployment.

Example:
```gotemplate
for _, d := range get_deployments_with_label_from_k8s(`proxless=true`) {
  uuid := get_uid_with_deployment_name(d.Name)
  route := get_route_with_uid(uid)
  if deployment_is_up(d) && is_deployment_idle(route) {
    scale_down(d)
  }
}
```

## Scalability

TODO

Each replica of proxless run the same code so they should normally have the same store at any time.  
However, the `last used` time is updated on only one of the replica at a time. We need a way to synchronise all the replicas with the latest `last used`. Therefore we will persist it in kubernetes.  

### Autosync 

Autosync run a separate routine. It is responsible for syncing the `last used` time to its corresponding proxless replica.
It will persist the `last used` time as an annotation on the service. Example: `proxless/last-used:[timestamp]`.  
Since we don't want to overwhelm the kubernetes API, autosync will only persist the timestamp every N seconds. `N` has to be chosen carefully since you should never be in a case when a replica downscale an app while another try to call it.  
Also, autosync will sync its own store with a latest `last use` if the data in kubernetes is more recent than one in its store.

Example:  
```gotemplate
for _, r := range get_routes_in_store() {
  deploy := get_deployment_from_k8s(r.deployName)
  if deploy.annotations[`proxless/last-used`] > r.lastUsed {
    r.lastUsed = deploy.annotations[`proxless/last-used`]
    update_store(r)
  } else {
    deploy.annotations[`proxless/last-used`] = r.lastUsed
    update_deployment(deploy)
  }
}
```