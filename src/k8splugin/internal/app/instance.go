/*
 * Copyright 2018 Intel Corporation, Inc
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
	"log"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// InstanceRequest contains the parameters needed for instantiation
// of profiles
type InstanceRequest struct {
	RBName      string            `json:"rb-name"`
	RBVersion   string            `json:"rb-version"`
	ProfileName string            `json:"profile-name"`
	CloudRegion string            `json:"cloud-region"`
	Labels      map[string]string `json:"labels"`
}

// InstanceResponse contains the response from instantiation
type InstanceResponse struct {
	ID        string                    `json:"id"`
	Request   InstanceRequest           `json:"request"`
	Namespace string                    `json:"namespace"`
	Resources []helm.KubernetesResource `json:"resources"`
}

// InstanceMiniResponse contains the response from instantiation
// It does NOT include the created resources.
// Use the regular GET to get the created resources for a particular instance
type InstanceMiniResponse struct {
	ID        string          `json:"id"`
	Request   InstanceRequest `json:"request"`
	Namespace string          `json:"namespace"`
}

// PodStatus defines the observed state of ResourceBundleState
type PodStatus struct {
	Name        string           `json:"name"`
	Namespace   string           `json:"namespace"`
	Ready       bool             `json:"ready"`
	Status      corev1.PodStatus `json:"status,omitempty"`
	IPAddresses []string         `json:"ipaddresses"`
}

// InstanceStatus is what is returned when status is queried for an instance
type InstanceStatus struct {
	Request         InstanceRequest  `json:"request"`
	Ready           bool             `json:"ready"`
	ResourceCount   int32            `json:"resourceCount"`
	PodStatuses     []PodStatus      `json:"podStatuses"`
	ServiceStatuses []corev1.Service `json:"serviceStatuses"`
}

// InstanceManager is an interface exposes the instantiation functionality
type InstanceManager interface {
	Create(i InstanceRequest) (InstanceResponse, error)
	Get(id string) (InstanceResponse, error)
	Status(id string) (InstanceStatus, error)
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
	storeName     string
	tagInst       string
	tagInstStatus string
}

// NewInstanceClient returns an instance of the InstanceClient
// which implements the InstanceManager
func NewInstanceClient() *InstanceClient {
	return &InstanceClient{
		storeName:     "rbdef",
		tagInst:       "instance",
		tagInstStatus: "instanceStatus",
	}
}

// Create an entry for the resource bundle profile in the database
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

	overrideValues := []string{}

	//Execute the kubernetes create command
	sortedTemplates, err := rb.NewProfileClient().Resolve(i.RBName, i.RBVersion, i.ProfileName, overrideValues)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Error resolving helm charts")
	}

	// TODO: Only generate if id is not provided
	id := namegenerator.Generate()

	k8sClient := KubernetesClient{}
	err = k8sClient.init(i.CloudRegion, id)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	createdResources, err := k8sClient.createResources(sortedTemplates, profile.Namespace)
	if err != nil {
		return InstanceResponse{}, pkgerrors.Wrap(err, "Create Kubernetes Resources")
	}

	//Compose the return response
	resp := InstanceResponse{
		ID:        id,
		Request:   i,
		Namespace: profile.Namespace,
		Resources: createdResources,
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

// Status returns the status for the instance
func (v *InstanceClient) Status(id string) (InstanceStatus, error) {

	//Read the status from the DB
	key := InstanceKey{
		ID: id,
	}

	value, err := db.DBconn.Read(v.storeName, key, v.tagInstStatus)
	if err != nil {
		return InstanceStatus{}, pkgerrors.Wrap(err, "Get Instance")
	}

	//value is a byte array
	if value != nil {
		resp := InstanceStatus{}
		err = db.DBconn.Unmarshal(value, &resp)
		if err != nil {
			return InstanceStatus{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
		}
		return resp, nil
	}

	return InstanceStatus{}, pkgerrors.New("Status is not available")
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
				ID:        resp.ID,
				Request:   resp.Request,
				Namespace: resp.Namespace,
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
	err = k8sClient.init(inst.Request.CloudRegion, inst.ID)
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
