package kube

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/logger"
	"time"
)

type kubeCluster struct {
	clientSet                      kubernetes.Interface
	servicesInformerResyncInterval int
}

func NewCluster(clientSet kubernetes.Interface) *kubeCluster {
	return &kubeCluster{
		clientSet:                      clientSet,
		servicesInformerResyncInterval: 60,
	}
}

func NewKubeClient(kubeConfigPath string) kubernetes.Interface {
	kubeConfigString := flag.String("kubeconfig", kubeConfigPath, "(optional) absolute path to the kubeconfig file")

	// use the current context in kubeconfig
	kubeConf, err := clientcmd.BuildConfigFromFlags("", *kubeConfigString)
	if err != nil {
		logger.Panicf(err, "Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return kubernetes.NewForConfigOrDie(kubeConf)
}

func (k *kubeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	return scaleUpDeployment(k.clientSet, name, namespace, timeout)
}

func (k *kubeCluster) ScaleDownDeployments(
	namespaceScope string, mustScaleDown func(deployName, namespace string) (bool, time.Duration, error)) []error {
	return scaleDownDeployments(k.clientSet, namespaceScope, mustScaleDown)
}

func (k *kubeCluster) RunServicesEngine(
	namespaceScope, proxlessService, proxlessNamespace string,
	upsertMemory func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromMemory func(id string) error,
) {
	// TODO make the resync is configurable
	runServicesInformer(
		k.clientSet, namespaceScope, proxlessService, proxlessNamespace, k.servicesInformerResyncInterval,
		upsertMemory, deleteRouteFromMemory)
}
