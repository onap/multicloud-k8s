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

package module

import (
	"encoding/json"
	"strings"

	jyaml "github.com/ghodss/yaml"

	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	pkgerrors "github.com/pkg/errors"
)

// NetControlIntent contains the parameters needed for dynamic networks
type NetControlIntent struct {
	Metadata Metadata `json:"metadata"`
}

// NetControlIntentKey is the key structure that is used in the database
type NetControlIntentKey struct {
	NetControlIntent    string `json:"netcontrolintent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
}

// Manager is an interface exposing the NetControlIntent functionality
type NetControlIntentManager interface {
	CreateNetControlIntent(nci NetControlIntent, project, compositeapp, compositeappversion string, exists bool) (NetControlIntent, error)
	GetNetControlIntent(name, project, compositeapp, compositeappversion string) (NetControlIntent, error)
	GetNetControlIntents(project, compositeapp, compositeappversion string) ([]NetControlIntent, error)
	DeleteNetControlIntent(name, project, compositeapp, compositeappversion string) error
	ApplyNetControlIntent(name, project, compositeapp, compositeappversion, appContextId string) error
}

// NetControlIntentClient implements the Manager
// It will also be used to maintain some localized state
type NetControlIntentClient struct {
	db ClientDbInfo
}

// NewNetControlIntentClient returns an instance of the NetControlIntentClient
// which implements the Manager
func NewNetControlIntentClient() *NetControlIntentClient {
	return &NetControlIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "netcontrolintentmetadata",
		},
	}
}

// CreateNetControlIntent - create a new NetControlIntent
func (v *NetControlIntentClient) CreateNetControlIntent(nci NetControlIntent, project, compositeapp, compositeappversion string, exists bool) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    nci.Metadata.Name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
	}

	//Check if this NetControlIntent already exists
	_, err := v.GetNetControlIntent(nci.Metadata.Name, project, compositeapp, compositeappversion)
	if err == nil && !exists {
		return NetControlIntent{}, pkgerrors.New("NetControlIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, nci)
	if err != nil {
		return NetControlIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return nci, nil
}

// GetNetControlIntent returns the NetControlIntent for corresponding name
func (v *NetControlIntentClient) GetNetControlIntent(name, project, compositeapp, compositeappversion string) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return NetControlIntent{}, pkgerrors.Wrap(err, "Get NetControlIntent")
	}

	//value is a byte array
	if value != nil {
		nci := NetControlIntent{}
		err = db.DBconn.Unmarshal(value[0], &nci)
		if err != nil {
			return NetControlIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return nci, nil
	}

	return NetControlIntent{}, pkgerrors.New("Error getting NetControlIntent")
}

// GetNetControlIntentList returns all of the NetControlIntent for corresponding name
func (v *NetControlIntentClient) GetNetControlIntents(project, compositeapp, compositeappversion string) ([]NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    "",
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
	}

	var resp []NetControlIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []NetControlIntent{}, pkgerrors.Wrap(err, "Get NetControlIntents")
	}

	for _, value := range values {
		nci := NetControlIntent{}
		err = db.DBconn.Unmarshal(value, &nci)
		if err != nil {
			return []NetControlIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, nci)
	}

	return resp, nil
}

// Delete the  NetControlIntent from database
func (v *NetControlIntentClient) DeleteNetControlIntent(name, project, compositeapp, compositeappversion string) error {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete NetControlIntent Entry;")
	}

	return nil
}

// (Test Routine) - Apply network-control-intent
func (v *NetControlIntentClient) ApplyNetControlIntent(name, project, compositeapp, compositeappversion, appContextId string) error {
	// TODO: Handle all Network Chain Intents for the Network Control Intent

	// Handle all Workload Intents for the Network Control Intent
	wis, err := NewWorkloadIntentClient().GetWorkloadIntents(project, compositeapp, compositeappversion, name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting Workload Intents for Network Control Intent %v for %v/%v%v not found", name, project, compositeapp, compositeappversion)
	}

	// Setup the AppContext
	var context appcontext.AppContext
	_, err = context.LoadAppContext(appContextId)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v for %v/%v%v",
			appContextId, project, compositeapp, compositeappversion)
	}

	// Handle all intents (currently just Interface intents) for each Workload Intent
	for _, wi := range wis {
		// The app/resource identified in the workload intent needs to be updated with two annotations.
		// 1 - The "k8s.v1.cni.cncf.io/networks" annotation will have {"name": "ovn-networkobj", "namespace": "default"} added
		//     to it (preserving any existing values for this annotation.
		// 2 - The "k8s.plugin.opnfv.org/nfn-network" annotation will add any network interfaces that are provided by the
		//     workload/interfaces intents.

		// Prepare the list of interfaces from the workload intent
		wifs, err := NewWorkloadIfIntentClient().GetWorkloadIfIntents(project,
			compositeapp,
			compositeappversion,
			name,
			wi.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err,
				"Error getting Workload Interface Intents for Workload Intent %v under Network Control Intent %v for %v/%v%v not found",
				wi.Metadata.Name, name, project, compositeapp, compositeappversion)
		}
		if len(wifs) == 0 {
			log.Warn("No interface intents provided for workload intent", log.Fields{
				"project":                project,
				"composite app":          compositeapp,
				"composite app version":  compositeappversion,
				"network control intent": name,
				"workload intent":        wi.Metadata.Name,
			})
			continue
		}

		// Get all clusters for the current App from the AppContext
		clusters, err := context.GetClusterNames(wi.Spec.AppName)
		for _, c := range clusters {
			rh, err := context.GetResourceHandle(wi.Spec.AppName, c,
				strings.Join([]string{wi.Spec.WorkloadResource, wi.Spec.Type}, "+"))
			if err != nil {
				log.Warn("App Context resource handle not found", log.Fields{
					"project":                project,
					"composite app":          compositeapp,
					"composite app version":  compositeappversion,
					"network control intent": name,
					"workload name":          wi.Metadata.Name,
					"app":                    wi.Spec.AppName,
					"resource":               wi.Spec.WorkloadResource,
					"resource type":          wi.Spec.Type,
				})
				continue
			}
			r, err := context.GetValue(rh)
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
			AddNetworkAnnotation(robj, netAnnot)

			// Add nfn interface annotations to object
			var newNfnIfs []WorkloadIfIntentSpec
			for _, i := range wifs {
				newNfnIfs = append(newNfnIfs, i.Spec)
			}
			AddNfnAnnotation(robj, newNfnIfs)

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
			err = context.UpdateResourceValue(rh, string(y))
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
