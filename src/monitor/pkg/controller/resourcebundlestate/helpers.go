package resourcebundlestate

import (
	"log"
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

// returnLabel verifies if the expected label exists and returns bool
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
