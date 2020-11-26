/*
=======================================================================
Copyright (c) 2017-2020 Aarna Networks, Inc.
All rights reserved.
======================================================================
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
          http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================================================================
*/

package app

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type GenericPlacementIntent struct {
	Metadata apiMetaData `json:"metadata"`
}

type PlacementIntent struct {
	Metadata apiMetaData            `json:"metadata"`
	Spec     AppPlacementIntentSpec `json:"spec"`
}
type PlacementIntentExport struct {
	Metadata apiMetaData            `json:"metadata"`
	Spec     AppPlacementIntentSpecExport `json:"spec"`
}

// appPlacementIntentSpec is the spec for per app intent
type AppPlacementIntentSpec struct {
	AppName string      `json:"app-name"`
	Intent  arrayIntent `json:"intent"`
}
type arrayIntent struct {
	AllofCluster []Allof `json:"allof"`
}
type Allof struct {
	ProviderName string `json:"provider-name"`
	ClusterName  string `json:"cluster-name"`
}
type AppPlacementIntentSpecExport struct {
	AppName string      `json:"appName"`
	Intent  arrayIntentExport `json:"intent"`
}
type arrayIntentExport struct {
	AllofCluster []AllofExport `json:"allof"`
}
type AllofExport struct {
	ProviderName string `json:"providerName"`
	ClusterName  string `json:"clusterName"`
}

// plamcentIntentHandler implements the orchworkflow interface
type placementIntentHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

type NetworkCtlIntent struct {
	Metadata apiMetaData `json:"metadata"`
}

type NetworkWlIntent struct {
	Metadata apiMetaData        `json:"metadata"`
	Spec     WorkloadIntentSpec `json:"spec"`
}

type WorkloadIntentSpec struct {
	AppName  string `json:"application-name"`
	Resource string `json:"workload-resource"`
	Type     string `json:"type"`
}
type WorkloadIntentSpecExport struct {
	AppName  string `json:"applicationName"`
	Resource string `json:"workloadResource"`
	Type     string `json:"type"`
}

type NwInterface struct {
	Metadata apiMetaData   `json:"metadata"`
	Spec     InterfaceSpec `json:"spec"`
}

type InterfaceSpec struct {
	Interface      string `json:"interface"`
	Name           string `json:"name"`
	DefaultGateway string `json:"defaultGateway"`
	IPAddress      string `json:"ipAddress"`
	MacAddress     string `json:"macAddress"`
}

// networkIntentHandler implements the orchworkflow interface
type networkIntentHandler struct {
	ovnURL       string
	orchInstance *OrchestrationHandler
}

func (h *placementIntentHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		Apps := compositeAppValue.AppsDataArray
		for digName, digValue := range Dig {
			h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName + "/generic-placement-intents"
			for gpintName, gpintValue := range digValue.GpintMap {
				for appName, _ := range Apps {
					var appPint PlacementIntent
					url := h.orchURL + "/" + gpintName + "/app-intents/" + appName + "_pint"
					retcode, retval, err := orch.apiGet(url, compositeAppMetadata.Name+"_getappPint")
					fmt.Printf("Get Gpint App intent in Composite app %s dig %s Gpint %s status %s\n",
						orch.compositeAppName, digName, gpintName, retcode)
					if err != nil {
						fmt.Printf("Failed to read app pint\n")
						return nil, 500
					}
					if retcode != 200 {
						fmt.Printf("Failed to read app pint\n")
						return nil, 200
					}
					err = json.Unmarshal(retval, &appPint)
					if err != nil {
						fmt.Printf("Failed to unmarshal json %s\n", err)
						return nil, 500
					}
					gpintValue.AppIntentArray = append(gpintValue.AppIntentArray, appPint)
				}
			}
		}
	}
	return nil, retcode
}

func (h *placementIntentHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			var gpintList []GenericPlacementIntent
			h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName + "/generic-placement-intents"
			retcode, retval, err := orch.apiGet(h.orchURL, compositeAppMetadata.Name+"_getgpint")
			fmt.Printf("Get Gpint in Composite app %s dig %s status %s\n", orch.compositeAppName,
				digName, retcode)
			if err != nil {
				fmt.Printf("Failed to read gpint\n")
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read gpint\n")
				return nil, retcode
			}
			json.Unmarshal(retval, &gpintList)
			digValue.GpintMap = make(map[string]*GpintData, len(gpintList))
			for _, value := range gpintList {
				var GpintDataInstance GpintData
				GpintDataInstance.Gpint = value
				digValue.GpintMap[value.Metadata.Name] = &GpintDataInstance
			}
		}
	}
	return nil, retcode
}

func (h *placementIntentHandler) deleteObject() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		Apps := compositeAppValue.AppsDataArray

		// loop through all app intens in the gpint
		for digName, digValue := range Dig {
			h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName + "/generic-placement-intents/"
			for gpintName, _ := range digValue.GpintMap {
				for appName, _ := range Apps {
					url := h.orchURL + gpintName +
						"/app-intents/" + appName + "_pint" // FIXME when query API works, change this API call to
					// query based on app name.
					fmt.Printf("Delete gping app intents %s\n", url)
					resp, err := orch.apiDel(url, orch.compositeAppName+"_delgpintintents")
					if err != nil {
						return err
					}
					if resp != 204 {
						return resp
					}
					fmt.Printf("Delete gpint intents resp %s\n", resp)
				}
			}
		}
	}
	return nil
}

func (h placementIntentHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap

		// loop through all app intens in the gpint
		for digName, digValue := range Dig {
			for gpintName, _ := range digValue.GpintMap {
				h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
					orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
					"/" + compositeAppSpec.Version +
					"/deployment-intent-groups/" + digName + "/generic-placement-intents/" +
					gpintName
				fmt.Printf("Delete gpint  %s\n", h.orchURL)
				resp, err := orch.apiDel(h.orchURL, compositeAppMetadata.Name+"_delgpints")
				if err != nil {
					return err
				}
				if resp != 204 {
					return resp
				}
				fmt.Printf("Delete gpint resp %s\n", resp)
			}
		}
	}
	return nil
}

func (h *placementIntentHandler) createAnchor() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData

	gpi := GenericPlacementIntent{
		Metadata: apiMetaData{
			Name:        intentData.CompositeAppName + "_gpint",
			Description: "Generic placement intent created from middleend",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}

	jsonLoad, _ := json.Marshal(gpi)
	// POST the generic placement intent
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + intentData.Spec.ProjectName +
		"/composite-apps/" + intentData.CompositeAppName + "/" + intentData.CompositeAppVersion +
		"/deployment-intent-groups/" + intentData.Name
	url := h.orchURL + "/generic-placement-intents"
	resp, err := orch.apiPost(jsonLoad, url, orch.digpIntents["generic-placement-intent"])
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	fmt.Printf("Generic placement intent resp %s\n", resp)

	return nil
}

func (h *placementIntentHandler) createObject() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData

	for _, value := range intentData.Spec.Apps {
		appName := value.Metadata.Name
		intentName := appName + "_pint"
		genericAppIntentName := intentData.CompositeAppName + "_gpint"
		providerName := value.Clusters[0].Provider
		clusterName := value.Clusters[0].SelectedClusters[0].Name

		pint := PlacementIntent{
			Metadata: apiMetaData{
				Name:        intentName,
				Description: "NA",
				UserData1:   "data 1",
				UserData2:   "data2"},
			Spec: AppPlacementIntentSpec{
				AppName: appName,
				Intent: arrayIntent{
					AllofCluster: []Allof{ // FIXME: the logic requires to handle allof/anyof and multi cluster.
						Allof{
							ProviderName: providerName,
							ClusterName:  clusterName},
					},
				},
			},
		}

		url := h.orchURL + "/generic-placement-intents/" + genericAppIntentName + "/app-intents"
		jsonLoad, _ := json.Marshal(pint)
		status, err := orch.apiPost(jsonLoad, url, intentName)
		if err != nil {
			log.Fatalln(err)
		}
		if status != 201 {
			return status
		}
		fmt.Printf("Placement intent %s status %s %s\n", intentName, status, url)
	}

	return nil
}

func addPlacementIntent(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.createObject()
	if err != nil {
		return err
	}
	return nil
}

func delGpint(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}

func (h *networkIntentHandler) createAnchor() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData

	nwIntent := NetworkCtlIntent{
		Metadata: apiMetaData{
			Name:        intentData.CompositeAppName + "_nwctlint",
			Description: "Network Controller created from middleend",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}
	jsonLoad, _ := json.Marshal(nwIntent)
	// POST the network controller intent
	h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" + intentData.Spec.ProjectName +
		"/composite-apps/" + intentData.CompositeAppName + "/" + intentData.CompositeAppVersion +
		"/deployment-intent-groups/" + intentData.Name
	url := h.ovnURL + "/network-controller-intent"
	resp, err := orch.apiPost(jsonLoad, url, orch.nwCtlIntents["network-controller-intent"])
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	fmt.Printf("Network contoller intent resp %s\n", resp)

	return nil
}

func (h *networkIntentHandler) createObject() interface{} {
	orch := h.orchInstance
	intentData := h.orchInstance.DigData

	for _, value := range intentData.Spec.Apps {

		appName := value.Metadata.Name
		intentName := value.Metadata.Name + "_wnwlint"
		genericAppIntentName := intentData.CompositeAppName + "_nwctlint"

		wlIntent := NetworkWlIntent{
			Metadata: apiMetaData{
				Name:        intentName,
				Description: "NA",
				UserData1:   "data 1",
				UserData2:   "data2"},
			Spec: WorkloadIntentSpec{
				AppName:  appName,
				Resource: appName,
				Type:     "deployment",
			},
		}

		url := h.ovnURL + "/network-controller-intent/" + genericAppIntentName + "/workload-intents"
		jsonLoad, _ := json.Marshal(wlIntent)
		status, err := orch.apiPost(jsonLoad, url, intentName)
		if err != nil {
			log.Fatalln(err)
		}
		if status != 201 {
			return status
		}
		fmt.Printf("Workload intent %s status %s %s\n", intentName, status, url)
	}

	// Add interfaces for to each application
	for _, value := range intentData.Spec.Apps {
		interfaces := value.Clusters[0].SelectedClusters[0].Interfaces
		for j := range interfaces {
			interfaceNum := strconv.Itoa(j)
			interfaceName := value.Metadata.Name + "_interface" + interfaceNum
			genericAppIntentName := intentData.CompositeAppName + "_nwctlint"
			workloadIntent := value.Metadata.Name + "_wnwlint"

			iface := NwInterface{
				Metadata: apiMetaData{
					Name:        interfaceName,
					Description: "NA",
					UserData1:   "data 1",
					UserData2:   "data2"},
				Spec: InterfaceSpec{
					Interface:      "eth" + interfaceNum,
					Name:           interfaces[j].NetworkName,
					DefaultGateway: "false",
					IPAddress:      interfaces[j].IP,
				},
			}

			url := h.ovnURL + "/network-controller-intent" + "/" + genericAppIntentName +
				"/workload-intents/" + workloadIntent + "/interfaces"
			jsonLoad, _ := json.Marshal(iface)
			status, err := orch.apiPost(jsonLoad, url, interfaceName)
			if err != nil {
				log.Fatalln(err)
			}
			if status != 201 {
				return status
			}
			fmt.Printf("interface %s status %s %s\n", interfaceName, status, url)
		}
	}

	return nil
}

func (h *networkIntentHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName
			for nwintName, nwintValue := range digValue.NwintMap {
				var wrlintList []NetworkWlIntent
				wlurl := h.ovnURL + "/network-controller-intent/" + nwintName + "/workload-intents"
				retcode, retval, err := orch.apiGet(wlurl, orch.compositeAppName+"_getnwwlint")
				fmt.Printf("Get Wrkld intents in Composite app %s dig %s nw intent %s status %d\n",
					orch.compositeAppName, digName, nwintName, retcode)
				if err != nil {
					fmt.Printf("Failed to read nw  workload int")
					return nil, 500
				}
				if retcode != 200 {
					fmt.Printf("Failed to read nw  workload int")
					return nil, retcode
				}
				json.Unmarshal(retval, &wrlintList)
				nwintValue.WrkintMap = make(map[string]*WrkintData, len(wrlintList))
				for _, wrlIntValue := range wrlintList {
					var WrkintDataInstance WrkintData
					WrkintDataInstance.Wrkint = wrlIntValue

					var ifaceList []NwInterface
					ifaceurl := h.ovnURL + "/network-controller-intent/" + nwintName +
						"/workload-intents/" + wrlIntValue.Metadata.Name + "/interfaces"
					retcode, retval, err := orch.apiGet(ifaceurl, orch.compositeAppName+"_getnwiface")
					fmt.Printf("Get interface in Composite app %s dig %s nw intent %s wrkld intent %s status %d\n",
						orch.compositeAppName, digName, nwintName, wrlIntValue.Metadata.Name, retcode)
					if err != nil {
						fmt.Printf("Failed to read nw interface")
						return nil, 500
					}
					if retcode != 200 {
						fmt.Printf("Failed to read nw interface")
						return nil, retcode
					}
					json.Unmarshal(retval, &ifaceList)
					WrkintDataInstance.Interfaces = ifaceList
					nwintValue.WrkintMap[wrlIntValue.Metadata.Name] = &WrkintDataInstance
				}
			}
		}
	}
	return nil, retcode
}

func (h *networkIntentHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName
			var nwintList []NetworkCtlIntent

			url := h.ovnURL + "/network-controller-intent"
			retcode, retval, err := orch.apiGet(url, orch.compositeAppName+"_getnwint")
			fmt.Printf("Get Network Ctl intent in Composite app %s dig %s status %d\n",
				orch.compositeAppName, digName, retcode)
			if err != nil {
				fmt.Printf("Failed to read nw int %s\n", err)
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read nw int")
				return nil, retcode
			}
			json.Unmarshal(retval, &nwintList)
			digValue.NwintMap = make(map[string]*NwintData, len(nwintList))
			for _, nwIntValue := range nwintList {
				var NwintDataInstance NwintData
				NwintDataInstance.Nwint = nwIntValue
				digValue.NwintMap[nwIntValue.Metadata.Name] = &NwintDataInstance
			}
		}
	}
	return nil, retcode
}

func (h *networkIntentHandler) deleteObject() interface{} {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName

			for nwintName, nwintValue := range digValue.NwintMap {
				for wrkintName, wrkintValue := range nwintValue.WrkintMap {
					// Delete the interfaces per workload intent.
					for _, value := range wrkintValue.Interfaces {
						url := h.ovnURL + "network-controller-intent/" + nwintName + "/workload-intents/" +
							wrkintName + "/interfaces/" + value.Spec.Name
						fmt.Printf("Delete app nw interface %s\n", url)
						retcode, err := orch.apiDel(url, orch.compositeAppName+"_delnwinterface")
						if err != nil {
							return err
						}
						if retcode != 204 {
							return retcode
						}
						fmt.Printf("Delete nw interface resp %s\n", retcode)
					}
					// Delete the workload intents.
					url := h.ovnURL + "network-controller-intent/" + nwintName + "/workload-intents/" + wrkintName
					fmt.Printf("Delete app nw wl intent %s\n", url)
					retcode, err := orch.apiDel(url, orch.compositeAppName+"_delnwwrkintent")
					if err != nil {
						return err
					}
					if retcode != 204 {
						return retcode
					}
					fmt.Printf("Delete nw wl intent resp %s\n", retcode)
				} // For workload intents in network controller intent.
			} // For network controller intents in Dig.
		} // For Dig.
	} // For composite app.
	return retcode
}

func (h networkIntentHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	retcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		Dig := compositeAppValue.DigMap
		for digName, digValue := range Dig {
			h.ovnURL = "http://" + orch.MiddleendConf.OvnService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + digName
			for nwintName, _ := range digValue.NwintMap {
				// loop through all app intens in the gpint
				url := h.ovnURL + "/network-controller-intent/" + nwintName
				fmt.Printf("Delete app nw controller intent %s\n", url)
				retcode, err := orch.apiDel(url, compositeAppMetadata.Name+"_delnwctlintent")
				if err != nil {
					return err
				}
				if retcode != 204 {
					return retcode
				}
				fmt.Printf("Delete nw controller intent %s\n", retcode)
			}
		}
	}
	return retcode
}

func addNetworkIntent(I orchWorkflow) interface{} {
	//1. Add network controller Intent
	err := I.createAnchor()
	if err != nil {
		return err
	}

	//2. Add network workload intent
	err = I.createObject()
	if err != nil {
		return err
	}

	return nil
}

func delNwintData(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}
