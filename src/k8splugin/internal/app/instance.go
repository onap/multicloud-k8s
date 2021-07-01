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
	"context"
	"encoding/json"
	"fmt"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"
	pkgerrors "github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	resource "k8s.io/cli-runtime/pkg/resource"
	"log"
	"strings"
	"time"
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
	Hooks       []*helm.Hook              `json:"-"`
}

// DbData contains the data to put to Db
type DbData struct {
	ID          string                    `json:"id"`
	Request     InstanceRequest           `json:"request"`
	Namespace   string                    `json:"namespace"`
	Status      string                    `json:"status"`
	ReleaseName string                    `json:"release-name"`
	Resources   []helm.KubernetesResource `json:"resources"`
	Hooks       []*helm.Hook              `json:"hooks"`
	HookProgress string					  `json:hook-progress`
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
	Status(id string) (InstanceStatus, error)
	Query(id, apiVersion, kind, name, labels string) (InstanceStatus, error)
	List(rbname, rbversion, profilename string) ([]InstanceMiniResponse, error)
	Find(rbName string, ver string, profile string, labelKeys map[string]string) ([]InstanceMiniResponse, error)
	Delete(id string) error
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
	readyChecker 	kube.ReadyChecker
}

// NewInstanceClient returns an instance of the InstanceClient
// which implements the InstanceManager
func NewInstanceClient() *InstanceClient {
	return &InstanceClient{
		storeName: "rbdef",
		tagInst:   "instance",
	}
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
	if i.OverrideValues != nil {
		for k, v := range i.OverrideValues {
			overrideValues = append(overrideValues, k+"="+v)
		}
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

	log.Printf("Hook info")
	for _,h := range hookList {
		log.Printf("  Name: %s", h.Hook.Name)
		log.Printf("    Events: %s", h.Hook.Events)
		log.Printf("    Weight: %d", h.Hook.Weight)
		log.Printf("    DeletePolicies: %s", h.Hook.DeletePolicies)
	}
	dbData := DbData{
		ID:                 id,
		Request:            i,
		Namespace:          profile.Namespace,
		ReleaseName:        releaseName,
		Status:             "PRE-INSTALL",
		Resources:          []helm.KubernetesResource{},
		Hooks:              hookList,
		HookProgress:		"",
	}

	key := InstanceKey{
		ID: id,
	}
	err = db.DBconn.Create(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	hookClient := NewHookClient(profile.Namespace, id)
	createdHooks, hooksToRun, err := hookClient.ExecHook(k8sClient, hookList, id, release.HookPreInstall, 60)
	dbData.HookProgress = fmt.Sprintf("%d/%d", len(createdHooks), len(hooksToRun))
	if err != nil {
		log.Printf("Error running preinstall hooks for release %s, Error: %s. Stop here", releaseName, err)
		dbData.Status = "PRE-INSTALL-FAILED"
		err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
		if err != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
		}
		return InstanceResponse{}, pkgerrors.Wrap(err, "Error running preinstall hooks")
	}

	dbData.Status = "CREATING"
	err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
	}
	createdResources, err := k8sClient.createResources(sortedTemplates, profile.Namespace);
	if err != nil {
		dbData.Status = "FAILED"
	} else {
		dbData.Status = "CREATED"
		dbData.Resources = createdResources
	}
	err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
	if err != nil {
		log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
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

	if dbData.Status == "FAILED" {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Error during instantiation of main rss, skip post-install hooks")
	}
	go func() {
		//No need to check for main rss status, based on the official implementation of helm
		//check if all rss are ready
		//for _, rss := range createdResources {
		//	isReady := false
		//	count := 0
		//	countMax := int(math.Floor(float64(1200)/5)) //timeout is 20m
		//	for !isReady {
		//		status, err := k8sClient.GetResourceStatus(rss, profile.Namespace);
		//		if err != nil {
		//			dbData.Status = "FAILED"
		//			log.Printf("  Release: %s, Get rss status %s has error, will not run the post-install hooks", releaseName, rss.Name)
		//			break
		//		}
		//		isReady, err = IsRssReady(rss.Name, status, true)
		//		if err != nil {
		//			dbData.Status = "FAILED"
		//			log.Printf("  Release: %s, Rss %s has error, will not run the post-install hooks", releaseName, rss.Name)
		//			break
		//		}
		//		if isReady {
		//			log.Printf("  Release: %s, creation complete for %s", releaseName, status.Name)
		//			break
		//		} else {
		//			log.Printf("  Release: %s, rss %s, kind %s is NOT READY, wait. couting left: %d", releaseName, rss.Name, rss.GVK.Kind, countMax - count)
		//			time.Sleep(5 * time.Second)
		//			if count++; count >= countMax {
		//				log.Printf("  Release: %s, warning: Rss %s cannot ready, consider deployment failed", releaseName, status.Name)
		//				dbData.Status = "FAILED"
		//				break
		//			}
		//		}
		//	}
		//}

		if dbData.Status != "FAILED" {
			dbData.Status = "POST-INSTALL"
			dbData.HookProgress = ""
		}
		db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
		if dbData.Status != "FAILED" {
			createdHooks, hooksToRun, err = hookClient.ExecHook(k8sClient, hookList, id, release.HookPostInstall, 600)
			dbData.HookProgress = fmt.Sprintf("%d/%d", len(createdHooks), len(hooksToRun))
			if err != nil {
				dbData.Status = "POST-INSTALL-FAILED"
				log.Printf("  Release: %s, Error running postinstall hooks error: %s", releaseName, err)
			} else {
				dbData.Status = "DONE"
			}
		} else {
			log.Printf("  Release: %s, Main rss are failed, skip post-install", releaseName)
		}
		err = db.DBconn.Update(v.storeName, key, v.tagInst, dbData)
		if err != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", releaseName)
		}
	}()

	return resp, nil
}

// Get returns the full instance for corresponding ID
func (v *InstanceClient) GetFull(id string) (DbData, error) {
	key := InstanceKey{
		ID: id,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagInst)
	if err != nil {
		return DbData{}, pkgerrors.Wrap(err, "Get Instance")
	}

	//value is a byte array
	if value != nil {
		resp := DbData{}
		err = db.DBconn.Unmarshal(value, &resp)
		if err != nil {
			return DbData{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
		}
		return resp, nil
	}

	return DbData{}, pkgerrors.New("Error getting Instance")
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

	dbData := DbData{}
	err = db.DBconn.Unmarshal(value, &dbData)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
	}

	k8sClient := KubernetesClient{}
	err = k8sClient.Init(dbData.Request.CloudRegion, id)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	cumulatedErrorMsg := make([]string, 0)
//	podsStatus, err := k8sClient.getPodsByLabel(resResp.Namespace)
//	if err != nil {
//		cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
//	}
//
//	generalStatus := make([]ResourceStatus, 0, len(resResp.Resources))
//Main:
//	for _, resource := range resResp.Resources {
//		for _, pod := range podsStatus {
//			if resource.GVK == pod.GVK && resource.Name == pod.Name {
//				continue Main //Don't double check pods if someone decided to define pod explicitly in helm chart
//			}
//		}
//		status, err := k8sClient.GetResourceStatus(resource, resResp.Namespace)
//		if err != nil {
//			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
//		} else {
//			generalStatus = append(generalStatus, status)
//		}
//	}
//	resp := InstanceStatus{
//		Request:         resResp.Request,
//		ResourceCount:   int32(len(generalStatus) + len(podsStatus)),
//		Ready:           false, //FIXME To determine readiness, some parsing of status fields is necessary
//		ResourcesStatus: append(generalStatus, podsStatus...),
//	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()
	isReady := true
	generalStatus := make([]ResourceStatus, 0, len(dbData.Resources))
	for _, oneResource := range dbData.Resources {
		apiVersion, kind := oneResource.GVK.ToAPIVersionAndKind()
		log.Printf("apiVersion: %s, Kind: %s", apiVersion, kind)
		restClient, err := k8sClient.getRestApi(apiVersion)

		if err != nil {
			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
			break
		}
		mapper := k8sClient.GetMapper()
		mapping, err := mapper.RESTMapping(schema.GroupKind{
			Group: oneResource.GVK.Group,
			Kind:  oneResource.GVK.Kind,
		}, oneResource.GVK.Version)
		resourceInfo := resource.Info{
			Client:          restClient,
			Mapping:         mapping,
			Namespace:       dbData.Namespace,
			Name:            oneResource.Name,
			Source:          "",
			Object:          nil,//we can keep nil here for all rss type except CRD because readyChecker.IsReady does not need this field
			ResourceVersion: "",
		}
		status, err := k8sClient.GetResourceStatus(oneResource, dbData.Namespace)
		if err != nil {
			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
		} else {
			generalStatus = append(generalStatus, status)
		}
		var parsedRes runtime.Object
		//TODO: Should we care about different api version for a same kind?
		switch kind {
		case "Pod":
			parsedRes = new(corev1.Pod)
		case "Job":
			parsedRes = new(batchv1.Job)
		case "Deployment":
			parsedRes = new(appsv1.Deployment)
		case "PersistentVolumeClaim":
			parsedRes = new(corev1.PersistentVolume)
		case "Service":
			parsedRes = new(corev1.Service)
		case "DaemonSet":
			parsedRes = new(appsv1.DaemonSet)
		case "CustomResourceDefinition":
			parsedRes = new(apiextv1.CustomResourceDefinition)
		case "StatefulSet":
			parsedRes = new(appsv1.StatefulSet)
		case "ReplicationController":
			parsedRes = new(corev1.ReplicationController)
		case "ReplicaSet":
			parsedRes = new(appsv1.ReplicaSet)
		default:
			//For not listed resource, consider ready
			continue
		}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(status.Status.Object, parsedRes)
		if err != nil {
			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
			break
		}
		resourceInfo.Object = parsedRes
		ready, err := k8sClient.readyChecker.IsReady(ctx, &resourceInfo)
		if !ready || err != nil {
			isReady = false
			break
		}
	}

	//resp := InstanceStatus{
	//	Request:         dbData.Request,
	//	ResourceCount:   int32(len(generalStatus) + len(podsStatus)),
	//	Ready:           false, //FIXME To determine readiness, some parsing of status fields is necessary
	//	ResourcesStatus: append(generalStatus, podsStatus...),
	//}
	resp := InstanceStatus{
		Request:         dbData.Request,
		ResourceCount:   int32(len(generalStatus)),
		Ready:           isReady,
		ResourcesStatus: generalStatus,
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

	hookClient := NewHookClient(inst.Namespace, id)
	createdHooks, hooksToRun, err := hookClient.ExecHook(k8sClient, inst.Hooks, id, release.HookPreDelete, 60)
	inst.HookProgress = fmt.Sprintf("%d/%d", len(createdHooks), len(hooksToRun))
	if err != nil {
		log.Printf("  Release: %s, Error running predelete hooks error: %s", inst.ReleaseName, err)
		inst.Status = "PRE-DELETE-FAILED"
		err = db.DBconn.Update(v.storeName, key, v.tagInst, inst)
		if err != nil {
			log.Printf("Update Instance DB Entry for release %s has error.", inst.ReleaseName)
		}
		return pkgerrors.Wrap(err, "Error running predelete hooks")
	}

	err = k8sClient.deleteResources(inst.Resources, inst.Namespace)
	if err != nil {
		return pkgerrors.Wrap(err, "Deleting Instance Resources")
	}

	go func() {
		//No need to check for main rss status, based on the official implementation of helm
		//for _, rss := range inst.Resources {
		//	isDeleted := false
		//	for !isDeleted {
		//		if _, err := k8sClient.GetResourceStatus(rss, inst.Namespace); err != nil {
		//			log.Printf("  Release: %s, rss %s is DELETED, continue", inst.ReleaseName, rss.Name)
		//			isDeleted = true
		//		} else {
		//			log.Printf("  Release: %s, rss %s is NOT DELETED, wait", inst.ReleaseName, rss.Name)
		//			time.Sleep(5 * time.Second)
		//		}
		//	}
		//}
		//
		//log.Printf("  Release: %s, All rss are deleted, continue", inst.ReleaseName)
		createdHooks, hooksToRun, err := hookClient.ExecHook(k8sClient, inst.Hooks, id, release.HookPostDelete, 600)
		inst.HookProgress = fmt.Sprintf("%d/%d", len(createdHooks), len(hooksToRun))
		if err != nil {
			//If this case happen, user should clean the cluster
			log.Printf("  Release: %s, Error running postdelete hooks error: %s", inst.ReleaseName, err)
			inst.Status = "POST-DELETE-FAILED"
			err = db.DBconn.Update(v.storeName, key, v.tagInst, inst)
			if err != nil {
				log.Printf("Update Instance DB Entry for release %s has error.", inst.ReleaseName)
			}
		}

		err = db.DBconn.Delete(v.storeName, key, v.tagInst)
		if err != nil {
			log.Printf("Delete Instance DB Entry for release %s has error.", inst.ReleaseName)
		}
	}()

	return nil
}
