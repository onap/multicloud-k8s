package main

import (
	"io/ioutil"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"k8-plugin-multicloud/krd"
)

func main() {}

// CreateResource object in a specific Kubernetes Deployment
func CreateResource(kubedata *krd.GenericKubeResourceData, kubeclient *kubernetes.Clientset) (string, error) {
	if kubedata.Namespace == "" {
		kubedata.Namespace = "default"
	}

	if _, err := os.Stat(kubedata.YamlFilePath); err != nil {
		return "", pkgerrors.New("File " + kubedata.YamlFilePath + " not found")
	}

	log.Println("Reading service YAML")
	rawBytes, err := ioutil.ReadFile(kubedata.YamlFilePath)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Service YAML file read error")
	}

	log.Println("Decoding service YAML")
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Deserialize service error")
	}

	switch o := obj.(type) {
	case *coreV1.Service:
		kubedata.ServiceData = o
	default:
		return "", pkgerrors.New(kubedata.YamlFilePath + " contains another resource different than Service")
	}

	kubedata.ServiceData.Namespace = kubedata.Namespace
	kubedata.ServiceData.Name = kubedata.InternalVNFID + "-" + kubedata.ServiceData.Name

	result, err := kubeclient.CoreV1().Services(kubedata.Namespace).Create(kubedata.ServiceData)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Service error")
	}
	return result.GetObjectMeta().GetName(), nil
}

// ListResources of existing deployments hosted in a specific Kubernetes Deployment
func ListResources(limit int64, namespace string, kubeclient *kubernetes.Clientset) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}
	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Service"

	list, err := kubeclient.CoreV1().Services(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Service list error")
	}
	result := make([]string, 0, limit)
	if list != nil {
		for _, service := range list.Items {
			result = append(result, service.Name)
		}
	}
	return &result, nil
}

// DeleteResource deletes an existing Kubernetes service
func DeleteResource(name string, namespace string, kubeclient *kubernetes.Clientset) error {
	if namespace == "" {
		namespace = "default"
	}

	log.Println("Deleting service: " + name)

	deletePolicy := metaV1.DeletePropagationForeground
	err := kubeclient.CoreV1().Services(namespace).Delete(name, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Service error")
	}

	return nil
}

// GetResource existing service hosting in a specific Kubernetes Service
func GetResource(name string, namespace string, kubeclient *kubernetes.Clientset) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: 10,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Service"

	list, err := kubeclient.CoreV1().Services(namespace).List(opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Deployment error")
	}

	for _, service := range list.Items {
		if name == service.Name {
			return name, nil
		}
	}

	return "", nil
}
