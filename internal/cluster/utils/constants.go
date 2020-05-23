package utils

const (
	LabelDeploymentProxless = "proxless"

	AnnotationServiceDomainKey               = "proxless/domains"
	AnnotationServiceDeployKey               = "proxless/deployment"
	AnnotationServiceTTLSeconds              = "proxless/ttl-seconds"
	AnnotationServiceReadinessTimeoutSeconds = "proxless/readiness-timeout-seconds"
)
