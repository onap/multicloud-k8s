/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package action

import (
	"encoding/json"
	"strings"

	jyaml "github.com/ghodss/yaml"

	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/ovnaction/pkg/module"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	pkgerrors "github.com/pkg/errors"
)

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextId)
	}
	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting metadata for AppContext with Id: %v", appContextId)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup

	// Handle all Workload Intents for the Network Control Intent
	wis, err := module.NewWorkloadIntentClient().GetWorkloadIntents(project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting Workload Intents for Network Control Intent %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	// Handle all intents (currently just Workload Interface intents) for each Workload Intent
	for _, wi := range wis {
		// The app/resource identified in the workload intent needs to be updated with two annotations.
		// 1 - The "k8s.v1.cni.cncf.io/networks" annotation will have {"name": "ovn-networkobj", "namespace": "default"} added
		//     to it (preserving any existing values for this annotation.
		// 2 - The "k8s.plugin.opnfv.org/nfn-network" annotation will add any network interfaces that are provided by the
		//     workload/interfaces intents.

		// Prepare the list of interfaces from the workload intent
		wifs, err := module.NewWorkloadIfIntentClient().GetWorkloadIfIntents(project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			wi.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err,
				"Error getting Workload Interface Intents for Workload Intent %v under Network Control Intent %v for %v/%v%v/%v not found",
				wi.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}
		if len(wifs) == 0 {
			log.Warn("No interface intents provided for workload intent", log.Fields{
				"project":                 project,
				"composite app":           compositeapp,
				"composite app version":   compositeappversion,
				"deployment intent group": deployIntentGroup,
				"network control intent":  intentName,
				"workload intent":         wi.Metadata.Name,
			})
			continue
		}

		// Get all clusters for the current App from the AppContext
		clusters, err := ac.GetClusterNames(wi.Spec.AppName)
		for _, c := range clusters {
			rh, err := ac.GetResourceHandle(wi.Spec.AppName, c,
				strings.Join([]string{wi.Spec.WorkloadResource, wi.Spec.Type}, "+"))
			if err != nil {
				log.Warn("App Context resource handle not found", log.Fields{
					"project":                 project,
					"composite app":           compositeapp,
					"composite app version":   compositeappversion,
					"deployment intent group": deployIntentGroup,
					"network control intent":  intentName,
					"workload name":           wi.Metadata.Name,
					"app":                     wi.Spec.AppName,
					"resource":                wi.Spec.WorkloadResource,
					"resource type":           wi.Spec.Type,
				})
				continue
			}
			r, err := ac.GetValue(rh)
			if err != nil {
				log.Error("Error retrieving resource from App Context", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
			}

			// Unmarshal resource to K8S object
			robj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r.(string)))

			// Add network annotation to object
			netAnnot := nettypes.NetworkSelectionElement{
				Name:      "ovn-networkobj",
				Namespace: "default",
			}
			module.AddNetworkAnnotation(robj, netAnnot)

			// Add nfn interface annotations to object
			var newNfnIfs []module.WorkloadIfIntentSpec
			for _, i := range wifs {
				newNfnIfs = append(newNfnIfs, i.Spec)
			}
			module.AddNfnAnnotation(robj, newNfnIfs)

			// Marshal object back to yaml format (via json - seems to eliminate most clutter)
			j, err := json.Marshal(robj)
			if err != nil {
				log.Error("Error marshalling resource to JSON", log.Fields{
					"error": err,
				})
				continue
			}
			y, err := jyaml.JSONToYAML(j)
			if err != nil {
				log.Error("Error marshalling resource to YAML", log.Fields{
					"error": err,
				})
				continue
			}

			// Update resource in AppContext
			err = ac.UpdateResourceValue(rh, string(y))
			if err != nil {
				log.Error("Network updating app context resource handle", log.Fields{
					"error":           err,
					"resource handle": rh,
				})
			}
		}
	}

	return nil
}
