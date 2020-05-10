# Deploy proxless using `kubectl`

## Proxless scoped

Proxless will only look at the services in the namespace where it is installed.

```shell script
kubectl apply -f proxless-scoped.yaml
```

## Proxless global

Proxless will look at all the services in all the namespaces.

_Notes: proxless will be installed in the `proxless` namespace_

```shell script
kubectl apply -f proxless-global.yaml
```