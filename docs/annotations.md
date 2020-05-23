# Annotations

Name | Description | Additional Information
--- | --- | ---
`proxless/domains` | comma separated list of domain names that will route to this service - the ingress must target proxless service |
`proxless/deployment` | name of the deployment associated to the service |
`proxless/ttl-seconds` | how many seconds proxless wait before scaling down the deployment when the service is not called | Optional - use env var `SERVERLESS_TTL_SECONDS` if empty
`proxless/readiness-timeout-seconds` | how much seconds proxless wait for the deployment to be ready when scaling up before timing out | Optional - use env var `DEPLOYMENT_READINESS_TIMEOUT_SECONDS` is empty

## Advanced use case

Adding the above annotations is enough for proxless to work correctly.  
As mentioned in [How Work Proxless](how-work-proxless.md), proxless will create a new service `[service-name]-proxless` for the internal connections.  
For example, the `hello-world` service will be accessible through `hello-world-proxless` from another kubernetes pod.

However, if you don't want proxless to manage the creation of the proxless service, you can create it by yourself.  
Therefore, this below annotation will be required on the service pointing to proxless.

/!\ It is important to add this annotation on the service poiting to proxless and not on the service poiting to the deployment.

Name | Description
--- | ---
`proxless/service` | Name of the service pointing to the deployment