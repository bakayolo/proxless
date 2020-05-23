package openshift

import (
	"github.com/openshift/client-go/apps/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/logger"
	"time"
)

type openshiftCluster struct {
	kubeClientSet                  kubernetes.Interface
	ocClientSet                    versioned.Interface
	servicesInformerResyncInterval int
}

func NewCluster(
	kubeClientSet kubernetes.Interface, ocClientSet versioned.Interface,
	servicesInformerResyncInterval int) cluster.Interface {
	return &openshiftCluster{
		kubeClientSet:                  kubeClientSet,
		ocClientSet:                    ocClientSet,
		servicesInformerResyncInterval: servicesInformerResyncInterval,
	}
}

func NewOpenshiftClient(kubeConfigPath string) versioned.Interface {
	// use the current context in kubeconfig
	kubeConf, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		logger.Panicf(err, "Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return versioned.NewForConfigOrDie(kubeConf)
}

func (o *openshiftCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	return scaleUpDeployment(o.ocClientSet, name, namespace, timeout)
}

func (o *openshiftCluster) ScaleDownDeployments(
	namespaceScope string, mustScaleDown func(deployName, namespace string) (bool, time.Duration, error)) []error {
	return scaleDownDeployments(o.ocClientSet, namespaceScope, mustScaleDown)
}

func (o *openshiftCluster) RunServicesEngine(
	namespaceScope, proxlessService, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	runServicesInformer(
		o.kubeClientSet, o.ocClientSet,
		namespaceScope, proxlessService, proxlessNamespace, o.servicesInformerResyncInterval,
		upsertMemory, deleteRouteFromMemory)
}
