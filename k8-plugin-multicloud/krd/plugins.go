package krd

import (
	"plugin"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// LoadedPlugins stores references to the stored plugins
var LoadedPlugins = map[string]*plugin.Plugin{}

// KubeResourceClient has the signature methods to create Kubernetes reources
type KubeResourceClient interface {
	CreateResource(GenericKubeResourceData, *kubernetes.Clientset) (string, error)
	ListResources(string, string) (*[]string, error)
	DeleteResource(string, string, *kubernetes.Clientset) error
	GetResource(string, string, *kubernetes.Clientset) (string, error)
}

// GenericKubeResourceData is a struct which stores all supported Kubernetes plugin types
type GenericKubeResourceData struct {
	YamlFilePath  string
	Namespace     string
	InternalVNFID string

	// Add additional Kubernetes plugins below kinds
	DeploymentData *appsV1.Deployment
	ServiceData    *coreV1.Service
}
