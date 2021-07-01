package app

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
)

func IsRssReady(name string, status ResourceStatus, isBeforePostInstall bool) (bool, error) {
	m, found, err := unstructured.NestedMap(status.Status.Object, "status")

	if err != nil {
		return false, err
	}
	if !found {
		if (status.GVK.Kind == "StatefulSet") || (status.GVK.Kind == "Job") || (status.GVK.Kind == "Deployment") || (status.GVK.Kind == "Pod") {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		log.Printf("    Rss %s status: %s", name, m)
		if status.GVK.Kind == "Service" {
			return true, nil
		} else if status.GVK.Kind == "StatefulSet" {
			if isBeforePostInstall {
				return true, nil
			}
			replicas, _, _ := unstructured.NestedInt64(m, "replicas")
			readyReplicas, _, _ := unstructured.NestedInt64(m, "readyReplicas")

			return readyReplicas == replicas, nil
		} else if status.GVK.Kind == "Pod" {
			if isBeforePostInstall {
				return true, nil
			}
			phase, found, err := unstructured.NestedString(m, "phase")
			if err != nil {
				return false, err
			}
			if found {
				if phase == "Running" {
					return true, err
				}
			}
			//no phase or phase is not Running -> check conditions
			conditions, found, err := unstructured.NestedSlice(m, "conditions")
			if err != nil {
				return false, err
			}
			if found {
				for _, oneCondition := range conditions {
					conditionType, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "type")
					//log.Printf("    Deployment status condition: condition: %s, type: %s", oneCondition, conditionType)
					if (conditionType == "Ready") || (conditionType == "ContainersReady") || (conditionType == "PodScheduled") {
						//have Ready, ContainersReady, Progressing or PodScheduled -> return true
						return true, nil
					}
				}
			} else {
				return false, nil
			}
		} else if status.GVK.Kind == "Deployment" {
			conditions, found, err := unstructured.NestedSlice(m, "conditions")
			if err != nil {
				return false, err
			}
			if found {
				for _, oneCondition := range conditions {
					conditionType, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "type")
					//log.Printf("    Deployment status condition: condition: %s, type: %s", oneCondition, conditionType)
					if (conditionType == "Ready") || (conditionType == "Available") || (conditionType == "Progressing") {
						//have Ready, ContainersReady, Progressing or PodScheduled -> return, no matter what
						return true, nil
					}
				}
			} else {
				return false, nil
			}
		} else if status.GVK.Kind == "Job" {
			if isBeforePostInstall {
				return true, nil
			}
			conditions, found, err := unstructured.NestedSlice(m, "conditions")
			if err != nil {
				return false, err
			}
			if found {
				for _, oneCondition := range conditions {
					conditionType, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "type")
					conditionStatus, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "status")
					//log.Printf("    Rss status condition: condition: %s, type: %s, status: %s", oneCondition, conditionType, conditionStatus)
					if ((conditionType == "Complete") || (conditionType == "Ready")) && (conditionStatus == "True") {
						return true, nil
					}
				}
			} else {
				return false, nil
			}
		} else if len(m) == 0 {
			//not all type above and have nothing in status field -> ready
			return true, nil
		} else {
			conditions, found, err := unstructured.NestedSlice(m, "conditions")
			if err != nil {
				return false, err
			}
			if found {
				for _, oneCondition := range conditions {
					conditionType, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "type")
					conditionStatus, _, _ := unstructured.NestedString(oneCondition.(map[string]interface{}), "status")
					//log.Printf("    Rss status condition: condition: %s, type: %s, status: %s", oneCondition, conditionType, conditionStatus)
					if ((conditionType == "Available") || (conditionType == "Ready")) && (conditionStatus == "True") {
						return true, nil
					}
				}
			} else {
				return true, nil
			}
		}
	}

	return false, nil
}
