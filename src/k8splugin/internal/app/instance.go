/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strings"

	pkgerrors "github.com/pkg/errors"
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
	storeName string
	tagInst   string
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

	createdResources, err := k8sClient.createResources(sortedTemplates, profile.Namespace)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Create Kubernetes Resources")
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

	key := InstanceKey{
		ID: id,
	}
	err = db.DBconn.Create(v.storeName, key, v.tagInst, resp)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	return resp, nil
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
	inst, err := v.Get(id)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting Instance")
	}

	k8sClient := KubernetesClient{}
	err = k8sClient.Init(inst.Request.CloudRegion, inst.ID)
	if err != nil {
		return pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	err = k8sClient.deleteResources(inst.Resources, inst.Namespace)
	if err != nil {
		return pkgerrors.Wrap(err, "Deleting Instance Resources")
	}

	key := InstanceKey{
		ID: id,
	}
	err = db.DBconn.Delete(v.storeName, key, v.tagInst)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Instance")
	}

	return nil
}
