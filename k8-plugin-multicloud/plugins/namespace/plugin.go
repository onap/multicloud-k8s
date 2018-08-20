package main

import (
	pkgerrors "github.com/pkg/errors"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {}

// CreateResource is used to create a new Namespace
func CreateResource(namespace string, client *kubernetes.Clientset) error {
	namespaceStruct := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(namespaceStruct)
	if err != nil {
		return pkgerrors.Wrap(err, "Create Namespace error")
	}
	return nil
}

// GetResource is used to check if a given namespace actually exists in Kubernetes
func GetResource(namespace string, client *kubernetes.Clientset) (bool, error) {
	ns, err := client.CoreV1().Namespaces().Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return false, pkgerrors.Wrap(err, "Get Namespace list error")
	}
	return ns != nil, nil
}

// DeleteResource is used to delete a namespace
func DeleteResource(namespace string, client *kubernetes.Clientset) error {
	deletePolicy := metaV1.DeletePropagationForeground

	err := client.CoreV1().Namespaces().Delete(namespace, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Namespace error")
	}
	return nil
}
