# Annotations

Name | Description | Additional Information
--- | --- | ---
`proxless/domains` | comma separated list of domain names that will route to this service - the ingress must target proxless service |
`proxless/deployment` | name of the deployment associated to the service |
`proxless/ttl-seconds` | how many seconds proxless wait before scaling down the deployment when the service is not called | Optional - use env var `SERVERLESS_TTL_SECONDS` if empty
`proxless/readiness-timeout-seconds` | how much seconds proxless wait for the deployment to be ready when scaling up before timing out | Optional - use env var `DEPLOYMENT_READINESS_TIMEOUT_SECONDS` is empty