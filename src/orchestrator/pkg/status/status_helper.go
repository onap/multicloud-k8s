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

package status

import (
	"encoding/json"
	"fmt"
	"strings"

	rb "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"github.com/onap/multicloud-k8s/src/monitor/pkg/generated/clientset/versioned/scheme"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/resourcestatus"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// decodeYAML reads a YAMl []byte to extract the Kubernetes object definition
func decodeYAML(y []byte, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(y, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

func getUnstruct(y []byte) (unstructured.Unstructured, error) {
	//Decode the yaml file to create a runtime.Object
	unstruct := unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := decodeYAML(y, &unstruct)
	if err != nil {
		log.Info(":: Error decoding YAML ::", log.Fields{"object": y, "error": err})
		return unstructured.Unstructured{}, pkgerrors.Wrap(err, "Decode object error")
	}

	return unstruct, nil
}

// GetClusterResources takes in a ResourceBundleStatus CR and resturns a list of ResourceStatus elments
func GetClusterResources(rbData rb.ResourceBundleStatus, qOutput string, qResources []string,
	resourceList *[]ResourceStatus, cnts map[string]int) (int, error) {

	count := 0

	for _, p := range rbData.PodStatuses {
		if !keepResource(p.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = p.Name
		r.Gvk = (&p.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = p
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.ServiceStatuses {
		if !keepResource(s.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, d := range rbData.DeploymentStatuses {
		if !keepResource(d.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, c := range rbData.ConfigMapStatuses {
		if !keepResource(c.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = c.Name
		r.Gvk = (&c.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = c
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.SecretStatuses {
		if !keepResource(s.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, d := range rbData.DaemonSetStatuses {
		if !keepResource(d.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = d.Name
		r.Gvk = (&d.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = d
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, i := range rbData.IngressStatuses {
		if !keepResource(i.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = i.Name
		r.Gvk = (&i.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = i
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, j := range rbData.JobStatuses {
		if !keepResource(j.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = j.Name
		r.Gvk = (&j.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = j
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	for _, s := range rbData.StatefulSetStatuses {
		if !keepResource(s.Name, qResources) {
			continue
		}
		r := ResourceStatus{}
		r.Name = s.Name
		r.Gvk = (&s.TypeMeta).GroupVersionKind()
		if qOutput == "detail" {
			r.Detail = s
		}
		*resourceList = append(*resourceList, r)
		count++
		cnt := cnts["Present"]
		cnts["Present"] = cnt + 1
	}

	return count, nil
}

// isResourceHandle takes a cluster handle and determines if the other handle parameter is a resource handle for this cluster
// handle.  It does this by verifying that the cluster handle is a prefix of the handle and that the remainder of the handle
// is a value that matches to a resource format:  "resource/<name>+<type>/"
// Example cluster handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/
// Example resource handle:
// /context/6385596659306465421/app/network-intents/cluster/vfw-cluster-provider+edge01/resource/emco-private-net+ProviderNetwork/
func isResourceHandle(ch, h interface{}) bool {
	clusterHandle := fmt.Sprintf("%v", ch)
	handle := fmt.Sprintf("%v", h)
	diff := strings.Split(handle, clusterHandle)

	if len(diff) != 2 && diff[0] != "" {
		return false
	}

	parts := strings.Split(diff[1], "/")

	if len(parts) == 3 &&
		parts[0] == "resource" &&
		len(strings.Split(parts[1], "+")) == 2 &&
		parts[2] == "" {
		return true
	} else {
		return false
	}
}

// keepResource keeps a resource if the filter list is empty or if the resource is part of the list
func keepResource(r string, rList []string) bool {
	if len(rList) == 0 {
		return true
	}
	for _, res := range rList {
		if r == res {
			return true
		}
	}
	return false
}

// GetAppContextResources collects the resource status of all resources in an AppContext subject to the filter parameters
func GetAppContextResources(ac appcontext.AppContext, ch interface{}, qOutput string, qResources []string, resourceList *[]ResourceStatus, statusCnts map[string]int) (int, error) {
	count := 0

	// Get all Resources for the Cluster
	hs, err := ac.GetAllHandles(ch)
	if err != nil {
		log.Info(":: Error getting all handles ::", log.Fields{"handles": ch, "error": err})
		return 0, err
	}

	for _, h := range hs {
		// skip any handles that are not resource handles
		if !isResourceHandle(ch, h) {
			continue
		}

		// Get Resource from AppContext
		res, err := ac.GetValue(h)
		if err != nil {
			log.Info(":: Error getting resource value ::", log.Fields{"Handle": h})
			continue
		}

		// Get Resource Status from AppContext
		sh, err := ac.GetLevelHandle(h, "status")
		if err != nil {
			log.Info(":: No status handle for resource ::", log.Fields{"Handle": h})
			continue
		}
		s, err := ac.GetValue(sh)
		if err != nil {
			log.Info(":: Error getting resource status value ::", log.Fields{"Handle": sh})
			continue
		}
		rstatus := resourcestatus.ResourceStatus{}
		js, err := json.Marshal(s)
		if err != nil {
			log.Info(":: Non-JSON status data for resource ::", log.Fields{"Handle": sh, "Value": s})
			continue
		}
		err = json.Unmarshal(js, &rstatus)
		if err != nil {
			log.Info(":: Invalid status data for resource ::", log.Fields{"Handle": sh, "Value": s})
			continue
		}

		// Get the unstructured object
		unstruct, err := getUnstruct([]byte(res.(string)))
		if err != nil {
			log.Info(":: Error getting GVK ::", log.Fields{"Resource": res, "error": err})
			continue
		}
		if !keepResource(unstruct.GetName(), qResources) {
			continue
		}

		// Make and fill out a ResourceStatus structure
		r := ResourceStatus{}
		r.Gvk = unstruct.GroupVersionKind()
		r.Name = unstruct.GetName()
		if qOutput == "detail" {
			r.Detail = unstruct.Object
		}
		r.RsyncStatus = fmt.Sprintf("%v", rstatus.Status)
		*resourceList = append(*resourceList, r)
		cnt := statusCnts[rstatus.Status]
		statusCnts[rstatus.Status] = cnt + 1
		count++
	}

	return count, nil
}

// PrepareStatusResult takes in a resource stateInfo object, the list of apps and the query parameters.
// It then fills out the StatusResult structure appropriately from information in the AppContext
func PrepareStatusResult(stateInfo state.StateInfo, apps []string, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (StatusResult, error) {

	var currentCtxId string
	if qInstance != "" {
		currentCtxId = qInstance
	} else {
		currentCtxId = state.GetLastContextIdFromStateInfo(stateInfo)
	}
	ac, err := state.GetAppContextFromId(currentCtxId)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext for status query not found")
	}

	// get the appcontext status value
	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext handle not found")
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status handle not found")
	}
	statusVal, err := ac.GetValue(sh)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "AppContext status value not found")
	}
	acStatus := appcontext.AppContextStatus{}
	js, err := json.Marshal(statusVal)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}
	err = json.Unmarshal(js, &acStatus)
	if err != nil {
		return StatusResult{}, pkgerrors.Wrap(err, "Invalid AppContext status value format")
	}

	statusResult := StatusResult{}

	statusResult.Apps = make([]AppStatus, 0)
	statusResult.State = stateInfo
	statusResult.Status = acStatus.Status

	rsyncStatusCnts := make(map[string]int)
	clusterStatusCnts := make(map[string]int)
	// Loop through each app and get the status data for each cluster in the app
	for _, app := range apps {
		appCount := 0
		if len(qApps) > 0 {
			found := false
			for _, a := range qApps {
				if a == app {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(app)
		if err != nil {
			continue
		}
		var appStatus AppStatus
		appStatus.Name = app
		appStatus.Clusters = make([]ClusterStatus, 0)

		for _, cluster := range clusters {
			clusterCount := 0
			if len(qClusters) > 0 {
				found := false
				for _, c := range qClusters {
					if c == cluster {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			var clusterStatus ClusterStatus
			pc := strings.Split(cluster, "+")
			clusterStatus.ClusterProvider = pc[0]
			clusterStatus.Cluster = pc[1]

			if qType == "cluster" {
				csh, err := ac.GetClusterStatusHandle(app, cluster)
				if err != nil {
					log.Info(":: No cluster status handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				clusterRbValue, err := ac.GetValue(csh)
				if err != nil {
					log.Info(":: No cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				var rbValue rb.ResourceBundleStatus
				err = json.Unmarshal([]byte(clusterRbValue.(string)), &rbValue)
				if err != nil {
					log.Info(":: Error unmarshaling cluster status value for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				clusterStatus.Resources = make([]ResourceStatus, 0)
				cnt, err := GetClusterResources(rbValue, qOutput, qResources, &clusterStatus.Resources, clusterStatusCnts)
				if err != nil {
					log.Info(":: Error gathering cluster resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				appCount += cnt
				clusterCount += cnt
			} else if qType == "rsync" {
				ch, err := ac.GetClusterHandle(app, cluster)
				if err != nil {
					log.Info(":: No handle for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}

				/* code to get status for resources from AppContext */
				clusterStatus.Resources = make([]ResourceStatus, 0)
				cnt, err := GetAppContextResources(ac, ch, qOutput, qResources, &clusterStatus.Resources, rsyncStatusCnts)
				if err != nil {
					log.Info(":: Error gathering appcontext resources for cluster, app ::",
						log.Fields{"Cluster": cluster, "AppName": app, "Error": err})
					continue
				}
				appCount += cnt
				clusterCount += cnt
			} else {
				log.Info(":: Invalid status type ::", log.Fields{"Status Type": qType})
				continue
			}

			if clusterCount > 0 {
				appStatus.Clusters = append(appStatus.Clusters, clusterStatus)
			}
		}
		if appCount > 0 && qOutput != "summary" {
			statusResult.Apps = append(statusResult.Apps, appStatus)
		}
	}
	statusResult.RsyncStatus = rsyncStatusCnts
	statusResult.ClusterStatus = clusterStatusCnts

	return statusResult, nil
}
