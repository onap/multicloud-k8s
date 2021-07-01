/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
 * Copyright © 2021 Nokia Bell Labs
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

package app

import (
	"encoding/json"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"
	pkgerrors "github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strconv"
	"strings"
)

// InstanceRequest contains the parameters needed for instantiation
// of profiles
type InstanceRequest struct {
	RBName         string            `json:"rb-name"`
	RBVersion      string            `json:"rb-version"`
	ProfileName    string            `json:"profile-name"`
	ReleaseName    string            `json:"release-name"`
	CloudRegion    string            `json:"cloud-region"`
	Labels         map[string]string `json:"labels"`
	OverrideValues map[string]string `json:"override-values"`
}

// InstanceResponse contains the response from instantiation
type InstanceResponse struct {
	ID          string                    `json:"id"`
	Request     InstanceRequest           `json:"request"`
	Namespace   string                    `json:"namespace"`
	ReleaseName string                    `json:"release-name"`
	Resources   []helm.KubernetesResource `json:"resources"`
	Hooks       []*helm.Hook              `json:"hooks"`
}

// InstanceDbData contains the data to put to Db
type InstanceDbData struct {
	ID          string                    `json:"id"`
	Request     InstanceRequest           `json:"request"`
	Namespace   string                    `json:"namespace"`
	Status      string                    `json:"status"`
	ReleaseName string                    `json:"release-name"`
	Resources   []helm.KubernetesResource `json:"resources"`
	Hooks       []*helm.Hook              `json:"hooks"`
	HookProgress string					  `json:"hook-progress"`
	PreInstallTimeout       int64      	  `json:"PreInstallTimeout"`
	PostInstallTimeout      int64         `json:"PostInstallTimeout"`
	PreDeleteTimeout        int64         `json:"PreDeleteTimeout"`
	PostDeleteTimeout       int64         `json:"PostDeleteTimeout"`
}

// InstanceMiniResponse contains the response from instantiation
// It does NOT include the created resources.
// Use the regular GET to get the created resources for a particular instance
type InstanceMiniResponse struct {
	ID          string          `json:"id"`
	Request     InstanceRequest `json:"request"`
	ReleaseName string          `json:"release-name"`
	Namespace   string          `json:"namespace"`
}

// InstanceStatus is what is returned when status is queried for an instance
type InstanceStatus struct {
	Request         InstanceRequest  `json:"request"`
	Ready           bool             `json:"ready"`
	ResourceCount   int32            `json:"resourceCount"`
	ResourcesStatus []ResourceStatus `json:"resourcesStatus"`
}

// InstanceManager is an interface exposes the instantiation functionality
type InstanceManager interface {
	Create(i InstanceRequest) (InstanceResponse, error)
	Get(id string) (InstanceResponse, error)
	GetFull(id string) (InstanceDbData, error)
	Status(id string) (InstanceStatus, error)
	Query(id, apiVersion, kind, name, labels string) (InstanceStatus, error)
	List(rbname, rbversion, profilename string) ([]InstanceMiniResponse, error)
	Find(rbName string, ver string, profile string, labelKeys map[string]string) ([]InstanceMiniResponse, error)
	Delete(id string) error
	RecoverCreateOrDelete(id string) error
}

// InstanceKey is used as the primary key in the db
type InstanceKey struct {
	ID string `json:"id"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk InstanceKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// InstanceClient implements the InstanceManager interface
// It will also be used to maintain some localized state
type InstanceClient struct {
	storeName 		string
	tagInst   		string
}

// NewInstanceClient returns an instance of the InstanceClient
// which implements the InstanceManager
func NewInstanceClient() *InstanceClient {
	newInstanceClient := &InstanceClient{
		storeName: "rbdef",
		tagInst:   "instance",
	}

	//Perform recovery here - Need to find a solution about data race error when test
	//Scan db and look for uncompleted tasks to continue
	//go func() {
	//	log.Printf("--------Recovering failed instances--------")
	//	resp, err := newInstanceClient.List("", "", "")
	//	if err != nil {
	//		if err.Error() != "Listing Instances: Did not find any objects with tag: instance" {
	//			log.Printf("Error listing instances, will not recover failed instances, error: " + err.Error())
	//		} else {
	//			log.Printf("--------DONE Recovering failed instances--------")
	//		}
	//		return
	//	}
	//	for _,instance := range resp {
	//		if err = newInstanceClient.RecoverCreateOrDelete(instance.ID); err!=nil {
	//			log.Printf(err.Error())
	//			continue
	//		}
	//	}
	//}()

	return newInstanceClient
}

// Simplified function to retrieve model data from instance ID
func resolveModelFromInstance(instanceID string) (rbName, rbVersion, profileName, releaseName string, err error) {
	v := NewInstanceClient()
	resp, err := v.Get(instanceID)
	if err != nil {
		return "", "", "", "", pkgerrors.Wrap(err, "Getting instance")
	}
	return resp.Request.RBName, resp.Request.RBVersion, resp.Request.ProfileName, resp.ReleaseName, nil
}

// Create an instance of rb on the cluster  in the database
func (v *InstanceClient) Create(i InstanceRequest) (InstanceResponse, error) {
	// Name is required
	if i.RBName == "" || i.RBVersion == "" || i.ProfileName == "" || i.CloudRegion == "" {
		return InstanceResponse{},
			pkgerrors.New("RBName, RBversion, ProfileName, CloudRegion are required to create a new instance")
	}

	//Check if profile exists
	profile, err := rb.NewProfileClient().Get(i.RBName, i.RBVersion, i.ProfileName)
	if err != nil {
		return InstanceResponse{}, pkgerrors.New("Unable to find Profile to create instance")
	}

	//Convert override values from map to array of strings of the following format
	//foo=bar
	overrideValues := []string{}
	var preInstallTimeOut, postInstallTimeOut, preDeleteTimeout, postDeleteTimeout int64
	if i.OverrideValues != nil {
		preInstallTimeOutStr, ok := i.OverrideValues["pre-install-timeout"]
		if !ok {
			preInstallTimeOutStr = "60"
		}
		preInstallTimeOut,err = strconv.ParseInt(preInstallTimeOutStr, 10, 64)
		if err != nil {
			return InstanceResponse{}, pkgerrors.Wrap(err, "Error parsing pre-install-timeout")
		}

		postInstallTimeOutStr, ok := i.OverrideValues["post-install-timeout"]
		if !ok {
			postInstallTimeOutStr = "600"
		}
		postInstallTimeOut,err = strconv.ParseInt(postInstallTimeOutStr, 10, 64)
		if err != nil {
			return InstanceResponse{}, pkgerrors.Wrap(err, "Error parsing post-install-timeout")
		}

		preDeleteTimeOutStr, ok := i.OverrideValues["pre-delete-timeout"]
		if !ok {
			preDeleteTimeOutStr = "60"
		}
		preDeleteTimeout,err = strconv.ParseInt(preDeleteTimeOutStr, 10, 64)
		if err != nil {
			return InstanceResponse{}, pkgerrors.Wrap(err, "Error parsing pre-delete-timeout")
		}

		postDeleteTimeOutStr, ok := i.OverrideValues["post-delete-timeout"]
		if !ok {
			postDeleteTimeOutStr = "600"
		}
		postDeleteTimeout,err = strconv.ParseInt(postDeleteTimeOutStr, 10, 64)
		if err != nil {
			return InstanceResponse{}, pkgerrors.Wrap(err, "Error parsing post-delete-timeout")
		}

		for k, v := range i.OverrideValues {
			overrideValues = append(overrideValues, k+"="+v)
		}
	} else {
		preInstallTimeOut = 60
		postInstallTimeOut = 600
		preDeleteTimeout = 60
		postDeleteTimeout = 600
	}

	//Execute the kubernetes create command
	sortedTemplates, hookList, releaseName, err := rb.NewProfileClient().Resolve(i.RBName, i.RBVersion, i.ProfileName, overrideValues, i.ReleaseName)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Error resolving helm charts")
	}

	// TODO: Only generate if id is not provided
	id := namegenerator.Generate()

	k8sClient := KubernetesClient{}
	err = k8sClient.Init(i.CloudRegion, id)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	log.Printf("Main rss info")
	for _,t := range sortedTemplates {
		log.Printf("  Path: %s", t.FilePath)
		log.Printf("    Kind: %s", t.GVK.Kind)
	}

	//Dry run the main rss to see if any of them are in the cluster already
	err = k8sClient.DryrunResources(sortedTemplates, profile.Namespace);
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Error when dryrun main rss")
	}

	log.Printf("Hook info")
	for _,h := range hookList {
		log.Printf("  Name: %s", h.Hook.Name)
		log.Printf("    Events: %s", h.Hook.Events)
		log.Printf("    Weight: %d", h.Hook.Weight)
		log.Printf("    DeletePolicies: %s", h.Hook.DeletePolicies)
	}
	dbData := InstanceDbData{
		ID:                 id,
		Request:            i,
		Namespace:          profile.Namespace,
		ReleaseName:        releaseName,
		Status:             "PRE-INSTALL",
		Resources:          []helm.KubernetesResource{},
		Hooks:              hookList,
		HookProgress:		"",
		PreInstallTimeout:  preInstallTimeOut,
		PostInstallTimeout: postInstallTimeOut,
		PreDeleteTimeout:   preDeleteTimeout,
		PostDeleteTimeout:  postDeleteTimeout,
	}

	key := InstanceKey{
		ID: id,
	}
	err = db.DBconn.Create(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	hookClient := NewHookClient(profile.Namespace, id, v.storeName, v.tagInst)
	if len(hookClient.getHookByEvent(hookList, release.HookPreInstall)) != 0 {
		err = hookClient.ExecHook(k8sClient, hookList, release.HookPreInstall, preInstallTimeOut, 0, &dbData)
		log.Printf("dbData.HookProgress %s", dbData.HookProgress)
		if err != nil {
			log.Printf("Error running preinstall hooks for release %s, Error: %s. Stop here", releaseName, err)
			dbData.Status = "PRE-INSTALL-FAILED"
			err2 := db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
			if err2 != nil {
				log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
			}
			return InstanceResponse{}, pkgerrors.Wrap(err, "Error running preinstall hooks")
		}
	}

	dbData.Status = "CREATING"
	err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Update Instance DB Entry")
	}

	//Main rss creation is supposed to be very quick -> no need to support recover for main rss
	createdResources, err := k8sClient.createResources(sortedTemplates, profile.Namespace);
	if err != nil {
		log.Printf("  Instance: %s, Main rss are failed, skip post-install and remove instance in DB", id)
		//main rss creation failed -> remove instance in DB
		err = db.DBconn.Delete(v.storeName, key, v.tagInst)
		if err != nil {
			log.Printf("Delete Instance DB Entry for release %s has error.", releaseName)
		}
		return InstanceResponse{}, pkgerrors.Wrap(err, "Create Kubernetes Resources")
	}

	dbData.Status = "CREATED"
	dbData.Resources = createdResources
	err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Update Instance DB Entry")
	}

	//Compose the return response
	resp := InstanceResponse{
		ID:          id,
		Request:     i,
		Namespace:   profile.Namespace,
		ReleaseName: releaseName,
		Resources:   createdResources,
		Hooks:       hookList,
	}

	if len(hookClient.getHookByEvent(hookList, release.HookPostInstall)) != 0 {
		go func() {
			dbData.Status = "POST-INSTALL"
			dbData.HookProgress = ""
			err = hookClient.ExecHook(k8sClient, hookList, release.HookPostInstall, postInstallTimeOut, 0, &dbData)
			log.Printf("dbData.HookProgress %s", dbData.HookProgress)
			if err != nil {
				dbData.Status = "POST-INSTALL-FAILED"
				log.Printf("  Instance: %s, Error running postinstall hooks error: %s", id, err)
			} else {
				dbData.Status = "DONE"
			}
			err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
			if err != nil {
				log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
			}
		}()
	} else {
		dbData.Status = "DONE"
		err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
		if err != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
		}
	}

	return resp, nil
}

// Get returns the full instance for corresponding ID
func (v *InstanceClient) GetFull(id string) (InstanceDbData, error) {
	key := InstanceKey{
		ID: id,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagInst)
	if err != nil {
		return InstanceDbData{}, pkgerrors.Wrap(err, "Get Instance")
	}

	//value is a byte array
	if value != nil {
		resp := InstanceDbData{}
		err = db.DBconn.Unmarshal(value, &resp)
		if err != nil {
			return InstanceDbData{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
		}
		return resp, nil
	}

	return InstanceDbData{}, pkgerrors.New("Error getting Instance")
}

// Get returns the instance for corresponding ID
func (v *InstanceClient) Get(id string) (InstanceResponse, error) {
	key := InstanceKey{
		ID: id,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagInst)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Get Instance")
	}

	//value is a byte array
	if value != nil {
		resp := InstanceResponse{}
		err = db.DBconn.Unmarshal(value, &resp)
		if err != nil {
			return InstanceResponse{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
		}
		//resp := InstanceResponse{
		//	ID:          dbData.ID,
		//	Request:     dbData.Request,
		//	Namespace:   dbData.Namespace,
		//	ReleaseName: dbData.ReleaseName,
		//	Resources:   dbData.Resources,
		//	Hooks:       dbData.Hooks,
		//}
		return resp, nil
	}

	return InstanceResponse{}, pkgerrors.New("Error getting Instance")
}

// Query returns state of instance's filtered resources
func (v *InstanceClient) Query(id, apiVersion, kind, name, labels string) (InstanceStatus, error) {

	//Read the status from the DB
	key := InstanceKey{
		ID: id,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagInst)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Get Instance")
	}
	if value == nil { //value is a byte array
		return InstanceStatus{}, pkgerrors.New("Status is not available")
	}
	resResp := InstanceResponse{}
	err = db.DBconn.Unmarshal(value, &resResp)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
	}

	k8sClient := KubernetesClient{}

	err = k8sClient.Init(resResp.Request.CloudRegion, id)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	var resourcesStatus []ResourceStatus
	if labels != "" {
		resList, err := k8sClient.queryResources(apiVersion, kind, labels, resResp.Namespace)
		if err != nil {
			return InstanceStatus{}, pkgerrors.Wrap(err, "Querying Resources")
		}
		// If user specifies both label and name, we want to pick up only single resource from these matching label
		if name != "" {
			//Assigning 0-length, because we may actually not find matching name
			resourcesStatus = make([]ResourceStatus, 0)
			for _, res := range resList {
				if res.Name == name {
					resourcesStatus = append(resourcesStatus, res)
					break
				}
			}
		} else {
			resourcesStatus = resList
		}
	} else if name != "" {
		resIdentifier := helm.KubernetesResource{
			Name: name,
			GVK:  schema.FromAPIVersionAndKind(apiVersion, kind),
		}
		res, err := k8sClient.GetResourceStatus(resIdentifier, resResp.Namespace)
		if err != nil {
			return InstanceStatus{}, pkgerrors.Wrap(err, "Querying Resource")
		}
		resourcesStatus = []ResourceStatus{res}
	}

	resp := InstanceStatus{
		Request:         resResp.Request,
		ResourceCount:   int32(len(resourcesStatus)),
		ResourcesStatus: resourcesStatus,
	}
	return resp, nil
}

// Status returns the status for the instance
func (v *InstanceClient) Status(id string) (InstanceStatus, error) {
	//TODO add a param to decide to add hook status or not in the response
	//Read the status from the DB
	key := InstanceKey{
		ID: id,
	}

	value, err := db.DBconn.Read(v.storeName, key, v.tagInst)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Get Instance")
	}

	//value is a byte array
	if value == nil {
		return InstanceStatus{}, pkgerrors.New("Status is not available")
	}

	resResp := InstanceResponse{}
	err = db.DBconn.Unmarshal(value, &resResp)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
	}

	k8sClient := KubernetesClient{}
	err = k8sClient.Init(resResp.Request.CloudRegion, id)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	cumulatedErrorMsg := make([]string, 0)
	podsStatus, err := k8sClient.getPodsByLabel(resResp.Namespace)
	if err != nil {
		cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
	}

	generalStatus := make([]ResourceStatus, 0, len(resResp.Resources))
Main:
	for _, resource := range resResp.Resources {
		for _, pod := range podsStatus {
			if resource.GVK == pod.GVK && resource.Name == pod.Name {
				continue Main //Don't double check pods if someone decided to define pod explicitly in helm chart
			}
		}
		status, err := k8sClient.GetResourceStatus(resource, resResp.Namespace)
		if err != nil {
			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
		} else {
			generalStatus = append(generalStatus, status)
		}
	}
	resp := InstanceStatus{
		Request:         resResp.Request,
		ResourceCount:   int32(len(generalStatus) + len(podsStatus)),
		Ready:           false, //FIXME To determine readiness, some parsing of status fields is necessary
		ResourcesStatus: append(generalStatus, podsStatus...),
	}

	if len(cumulatedErrorMsg) != 0 {
		err = pkgerrors.New("Getting Resources Status:\n" +
			strings.Join(cumulatedErrorMsg, "\n"))
		return resp, err
	}
	//TODO Filter response content by requested verbosity (brief, ...)?

	return resp, nil
}

// List returns the instance for corresponding ID
// Empty string returns all
func (v *InstanceClient) List(rbname, rbversion, profilename string) ([]InstanceMiniResponse, error) {

	dbres, err := db.DBconn.ReadAll(v.storeName, v.tagInst)
	if err != nil || len(dbres) == 0 {
		return []InstanceMiniResponse{}, pkgerrors.Wrap(err, "Listing Instances")
	}

	var results []InstanceMiniResponse

	for key, value := range dbres {
		//value is a byte array
		if value != nil {
			resp := InstanceResponse{}
			err = db.DBconn.Unmarshal(value, &resp)
			if err != nil {
				log.Printf("[Instance] Error: %s Unmarshaling Instance: %s", err.Error(), key)
			}

			miniresp := InstanceMiniResponse{
				ID:          resp.ID,
				Request:     resp.Request,
				Namespace:   resp.Namespace,
				ReleaseName: resp.ReleaseName,
			}

			//Filter based on the accepted keys
			if len(rbname) != 0 &&
				miniresp.Request.RBName != rbname {
				continue
			}
			if len(rbversion) != 0 &&
				miniresp.Request.RBVersion != rbversion {
				continue
			}
			if len(profilename) != 0 &&
				miniresp.Request.ProfileName != profilename {
				continue
			}

			results = append(results, miniresp)
		}
	}

	return results, nil
}

// Find returns the instances that match the given criteria
// If version is empty, it will return all instances for a given rbName
// If profile is empty, it will return all instances for a given rbName+version
// If labelKeys are provided, the results are filtered based on that.
// It is an AND operation for labelkeys.
func (v *InstanceClient) Find(rbName string, version string, profile string, labelKeys map[string]string) ([]InstanceMiniResponse, error) {
	if rbName == "" && len(labelKeys) == 0 {
		return []InstanceMiniResponse{}, pkgerrors.New("rbName or labelkeys is required and cannot be empty")
	}

	responses, err := v.List(rbName, version, profile)
	if err != nil {
		return []InstanceMiniResponse{}, pkgerrors.Wrap(err, "Listing Instances")
	}

	ret := []InstanceMiniResponse{}

	//filter the list by labelKeys now
	for _, resp := range responses {

		add := true
		for k, v := range labelKeys {
			if resp.Request.Labels[k] != v {
				add = false
				break
			}
		}
		// If label was not found in the response, don't add it
		if add {
			ret = append(ret, resp)
		}
	}

	return ret, nil
}

// Delete the Instance from database
func (v *InstanceClient) Delete(id string) error {
	inst, err := v.GetFull(id)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting Instance")
	}
	key := InstanceKey{
		ID: id,
	}
	if inst.Status == "DELETED" {
		//The instance is deleted when the plugin comes back -> just remove from Db
		err = db.DBconn.Delete(v.storeName, key, v.tagInst)
		if err != nil {
			log.Printf("Delete Instance DB Entry for release %s has error.", inst.ReleaseName)
		}
		return nil
	} else if inst.Status != "DONE"{
		//Recover is ongoing, do nothing here
		return nil
	}

	k8sClient := KubernetesClient{}
	err = k8sClient.Init(inst.Request.CloudRegion, inst.ID)
	if err != nil {
		return pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}
	inst.Status = "PRE-DELETE"
	inst.HookProgress = ""
	err = db.DBconn.Update(v.storeName, key, v.tagInst, inst)
	if err != nil {
		log.Printf("Update Instance DB Entry for release %s has error.", inst.ReleaseName)
	}

	hookClient := NewHookClient(inst.Namespace, id, v.storeName, v.tagInst)
	if len(hookClient.getHookByEvent(inst.Hooks, release.HookPreDelete)) != 0 {
		err = hookClient.ExecHook(k8sClient, inst.Hooks, release.HookPreDelete, inst.PreDeleteTimeout, 0, &inst)
		if err != nil {
			log.Printf("  Instance: %s, Error running pre-delete hooks error: %s", id, err)
			inst.Status = "PRE-DELETE-FAILED"
			err2 := db.DBconn.Update(v.storeName, key, v.tagInst, inst)
			if err2 != nil {
				log.Printf("Update Instance DB Entry for release %s has error.", inst.ReleaseName)
			}
			return pkgerrors.Wrap(err, "Error running pre-delete hooks")
		}
	}


	inst.Status = "DELETING"
	err = db.DBconn.Update(v.storeName, key, v.tagInst, inst)
	if err != nil {
		log.Printf("Update Instance DB Entry for release %s has error.", inst.ReleaseName)
	}
	err = k8sClient.deleteResources(inst.Resources, inst.Namespace)
	if err != nil {
		return pkgerrors.Wrap(err, "Deleting Instance Resources")
	}
	if len(hookClient.getHookByEvent(inst.Hooks, release.HookPostDelete)) != 0 {
		go func() {
			inst.HookProgress = ""
			if err := v.runPostDelete(k8sClient, hookClient, &inst, 0, true); err != nil {
				log.Printf(err.Error())
			}
		}()
	} else {
		err = db.DBconn.Delete(v.storeName, key, v.tagInst)
		if err != nil {
			return pkgerrors.Wrap(err, "Delete Instance")
		}
	}

	return nil
}

//Continue the instantiation
func (v *InstanceClient) RecoverCreateOrDelete(id string) error {
	instance, err := v.GetFull(id)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting instance " + id + ", skip this instance.  Error detail")
	}
	log.Printf("Instance " + id + ", status: " + instance.Status + ", HookProgress: " + instance.HookProgress)
	//have to resolve again template for this instance because all templates are in /tmp -> will be deleted when container restarts
	overrideValues := []string{}
	if instance.Request.OverrideValues != nil {
		for k, v := range instance.Request.OverrideValues {
			overrideValues = append(overrideValues, k + "=" + v)
		}
	}
	key := InstanceKey{
		ID: id,
	}
	log.Printf("  Resolving template for release %s", instance.Request.ReleaseName)
	_, hookList, _, err := rb.NewProfileClient().Resolve(instance.Request.RBName, instance.Request.RBVersion, instance.Request.ProfileName, overrideValues, instance.Request.ReleaseName)
	instance.Hooks = hookList
	err = db.DBconn.Update(v.storeName, key, v.tagInst, instance)
	if err != nil {
		return pkgerrors.Wrap(err, "Update Instance DB Entry")
	}

	if strings.Contains(instance.Status, "FAILED"){
		log.Printf("  This instance has failed during instantiation, not going to recover")
		return nil
	} else if !strings.Contains(instance.Status, "-INSTALL") && !strings.Contains(instance.Status, "-DELETE") {
		log.Printf("  This instance is not in hook state, not going to recover")
		return nil
	}

	splitHookProgress := strings.Split(instance.HookProgress,"/")
	completedHooks,err := strconv.Atoi(splitHookProgress[0])
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting completed PRE-INSTALL hooks for instance  " + instance.ID + ", skip. Error detail")
	}

	//we can add an option to delete instances that will not be recovered from database to clean the db
	if (instance.Status != "POST-INSTALL") && (instance.Status != "PRE-DELETE") && (instance.Status != "POST-DELETE") {
		if instance.Status == "PRE-INSTALL" {
			//Plugin quits during pre-install hooks -> Will do nothing because from SO point of view, there's no instance ID and will be reported as fail and be rolled back
			log.Printf("  The plugin quits during pre-install hook of this instance, not going to recover")
		}
		return nil
	}
	k8sClient := KubernetesClient{}
	err = k8sClient.Init(instance.Request.CloudRegion, id)
	if err != nil {
		log.Printf("  Error getting CloudRegion %s", instance.Request.CloudRegion)
		return nil
	}
	hookClient := NewHookClient(instance.Namespace, id, v.storeName, v.tagInst)
	switch instance.Status {
	case "POST-INSTALL":
		//Plugin quits during post-install hooks -> continue
		go func() {
			log.Printf("  The plugin quits during post-install hook of this instance, continue post-install hook")
			err = hookClient.ExecHook(k8sClient, instance.Hooks, release.HookPostInstall, instance.PostInstallTimeout, completedHooks, &instance)
			log.Printf("dbData.HookProgress %s", instance.HookProgress)
			if err != nil {
				instance.Status = "POST-INSTALL-FAILED"
				log.Printf("  Instance: %s, Error running postinstall hooks error: %s", id, err)
			} else {
				instance.Status = "DONE"
			}
			err = db.DBconn.Update(v.storeName, key, v.tagInst, instance)
			if err != nil {
				log.Printf("Update Instance DB Entry for release %s has error.", instance.ReleaseName)
			}
		}()
	case "PRE-DELETE":
		//Plugin quits during pre-delete hooks -> This already effects the instance -> should continue the deletion
		go func() {
			log.Printf("  The plugin quits during pre-delete hook of this instance, continue pre-delete hook")
			err = hookClient.ExecHook(k8sClient, instance.Hooks, release.HookPreDelete, instance.PreDeleteTimeout, completedHooks, &instance)
			if err != nil {
				log.Printf("  Instance: %s, Error running pre-delete hooks error: %s", id, err)
				instance.Status = "PRE-DELETE-FAILED"
				err = db.DBconn.Update(v.storeName, key, v.tagInst, instance)
				if err != nil {
					log.Printf("Update Instance DB Entry for release %s has error.", instance.ReleaseName)
				}
				return
			}

			err = k8sClient.deleteResources(instance.Resources, instance.Namespace)
			if err != nil {
				log.Printf("  Error running deleting instance resources, error: %s", err)
				return
			}
			//will not delete the instance in Db to avoid error when SO call delete again and there is not instance in DB
			//the instance in DB will be deleted when SO call delete again.
			instance.HookProgress = ""
			if err := v.runPostDelete(k8sClient, hookClient, &instance, 0, false); err != nil {
				log.Printf(err.Error())
			}
		}()
	case "POST-DELETE":
		//Plugin quits during post-delete hooks -> continue
		go func() {
			log.Printf("  The plugin quits during post-delete hook of this instance, continue post-delete hook")
			if err := v.runPostDelete(k8sClient, hookClient, &instance, completedHooks, true); err != nil {
				log.Printf(err.Error())
			}
		}()
	default:
		log.Printf("  This instance is not in hook state, not going to recover")
	}

	return nil
}

func (v *InstanceClient) runPostDelete(k8sClient KubernetesClient,hookClient *HookClient, instance *InstanceDbData, startIndex int, clearDb bool) error {
	key := InstanceKey{
		ID: instance.ID,
	}
	instance.Status = "POST-DELETE"
	err := db.DBconn.Update(v.storeName, key, v.tagInst, instance)
	if err != nil {
		log.Printf("Update Instance DB Entry for release %s has error.", instance.ReleaseName)
	}
	err = hookClient.ExecHook(k8sClient, instance.Hooks, release.HookPostDelete, instance.PostDeleteTimeout, startIndex, instance)
	if err != nil {
		//If this case happen, user should clean the cluster
		log.Printf("  Instance: %s, Error running post-delete hooks error: %s", instance.ID, err)
		instance.Status = "POST-DELETE-FAILED"
		err2 := db.DBconn.Update(v.storeName, key, v.tagInst, instance)
		if err2 != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", instance.ReleaseName)
			return pkgerrors.Wrap(err2, "Error running post-delete hooks")
		}
		return pkgerrors.Wrap(err, "Error running post-delete hooks")
	}
	if clearDb {
		err = db.DBconn.Delete(v.storeName, key, v.tagInst)
		if err != nil {
			log.Printf("Delete Instance DB Entry for release %s has error.", instance.ReleaseName)
		}
	} else {
		instance.Status = "DELETED"
		err2 := db.DBconn.Update(v.storeName, key, v.tagInst, instance)
		if err2 != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", instance.ReleaseName)
		}
	}
	return nil
}
