package main

import (
	"k8s.io/client-go/kubernetes"

	"github.com/shank7485/k8-plugin-multicloud/krd"
)

func main() {}

// CreateResource object in a specific Kubernetes resource
func CreateResource(kubedata *krd.GenericKubeResourceData, kubeclient *kubernetes.Clientset) (string, error) {
	return "externalUUID", nil
}

// ListResources of existing resources
func ListResources(limit int64, namespace string, kubeclient *kubernetes.Clientset) (*[]string, error) {
	returnVal := []string{"cloud1-default-uuid1", "cloud1-default-uuid2"}
	return &returnVal, nil
}

// DeleteResource existing resources
func DeleteResource(name string, namespace string, kubeclient *kubernetes.Clientset) error {
	return nil
}

// GetResource existing resource host
func GetResource(namespace string, client *kubernetes.Clientset) (bool, error) {
	return true, nil
}
