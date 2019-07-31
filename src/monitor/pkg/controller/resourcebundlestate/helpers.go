package resourcebundlestate

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkLabel verifies if the expected label exists and returns bool
func checkLabel(labels map[string]string) bool {

	_, ok := labels["k8splugin.io/rb-inst-id"]
	if !ok {
		log.Printf("Pod does not have label. Filter it.")
		return false
	}
	return true
}

// returnLabel verifies if the expected label exists and returns a map
func returnLabel(labels map[string]string) map[string]string {

	l, ok := labels["k8splugin.io/rb-inst-id"]
	if !ok {
		log.Printf("Pod does not have label. Filter it.")
		return nil
	}
	return map[string]string{
		"k8splugin.io/rb-inst-id": l,
	}
}

// listResources lists resources based on the selectors provided
// The data is returned in the pointer to the runtime.Object
// provided as argument.
func listResources(cli client.Client, namespace string,
	labelSelector map[string]string, returnData runtime.Object) error {

	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(labelSelector),
	}

	err := cli.List(context.TODO(), listOptions, returnData)
	if err != nil {
		log.Printf("Failed to list CRs: %v", err)
		return err
	}

	return nil
}
