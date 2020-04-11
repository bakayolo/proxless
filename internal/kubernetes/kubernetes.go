package kubernetes

import (
	"flag"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kube-proxless/internal/config"
)

var (
	ClientSet *kubernetes.Clientset
)

func InitKubeClient() {
	kubeConf := loadKubeConfig(config.KubeConfigPath)
	ClientSet = kubernetes.NewForConfigOrDie(kubeConf)
}

func loadKubeConfig(kubeConfigPath string) *rest.Config {
	kubeConfigString := flag.String("kubeconfig", kubeConfigPath, "(optional) absolute path to the kubeconfig file")

	// use the current context in kubeconfig
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeConfigString)
	if err != nil {
		log.Panic().Err(err).Msgf("Could not find kubeconfig file at %s", kubeConfigPath)
	}

	return kubeConfig
}
