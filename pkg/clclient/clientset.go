package clclient

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	var config *rest.Config

	if info, _ := os.Stat(kubeconfig); info == nil {
		var err error

		log.Printf("Using in-cluster Kubernetes API client")

		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, xerrors.Errorf("GetNodeSClient: %w", err)
		}
	} else {
		var err error

		log.Printf("Using kubeconfig-based Kubernetes API client")

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, xerrors.Errorf("GetNodesClient: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, xerrors.Errorf("new for config: %w", err)
	}

	return clientset, nil
}
