package kube

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/logger"
)

type kubeCluster struct {
	clientSet                      kubernetes.Interface
	servicesInformerResyncInterval int
}

func NewCluster(clientSet kubernetes.Interface, servicesInformerResyncInterval int) cluster.Interface {
	return &kubeCluster{
		clientSet:                      clientSet,
		servicesInformerResyncInterval: servicesInformerResyncInterval,
	}
}

func NewKubeClient(kubeConfigPath string) kubernetes.Interface {
	// use the current context in kubeconfig
	kubeConf, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		logger.Panicf(err, "Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return kubernetes.NewForConfigOrDie(kubeConf)
}

func (k *kubeCluster) ScaleUpDeployment(name, namespace string, timeout int) error {
	return scaleUpDeployment(k.clientSet, name, namespace, timeout)
}

func (k *kubeCluster) ScaleDownDeployment(deploymentName, namespace string) error {
	return scaleDownDeployment(k.clientSet, deploymentName, namespace)
}

func (k *kubeCluster) RunServicesEngine(
	namespaceScope, proxlessService, proxlessNamespace string,
	upsertMemory func(
		id, name, port, deployName, namespace string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) error,
	deleteRouteFromMemory func(id string) error,
) {
	runServicesInformer(
		k.clientSet, namespaceScope, proxlessService, proxlessNamespace, k.servicesInformerResyncInterval,
		upsertMemory, deleteRouteFromMemory)
}
