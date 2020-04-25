package kube

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/logger"
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

func (c *kubeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	return scaleUpDeployment(c.clientSet, name, namespace, timeout)
}

func (c *kubeCluster) ScaleDownDeployments(
	namespace string, mustScaleDown func(deployName, namespace string) (bool, error)) []error {
	return scaleDownDeployments(c.clientSet, namespace, mustScaleDown)
}

func (c *kubeCluster) RunServicesEngine(
	namespace string,
	upsertStore func(id, name, port, deployName, namespace string, domains []string) error,
	deleteRouteFromStore func(id string) error,
) {
	// TODO make the resync is configurable
	runServicesInformer(c.clientSet, namespace, c.servicesInformerResyncInterval, upsertStore, deleteRouteFromStore)
}
