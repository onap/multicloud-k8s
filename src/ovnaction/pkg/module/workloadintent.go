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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// WorkloadIntent contains the parameters needed for dynamic networks
type WorkloadIntent struct {
	Metadata Metadata           `json:"metadata"`
	Spec     WorkloadIntentSpec `json:"spec"`
}

type WorkloadIntentSpec struct {
	AppName          string `json:"application-name"`
	WorkloadResource string `json:"workload-resource"`
	Type             string `json:"type"`
}

// WorkloadIntentKey is the key structure that is used in the database
type WorkloadIntentKey struct {
	Project             string `json:"provider"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	WorkloadIntent      string `json:"workloadintent"`
}

// Manager is an interface exposing the WorkloadIntent functionality
type WorkloadIntentManager interface {
	CreateWorkloadIntent(wi WorkloadIntent, project, compositeapp, compositeappversion, dig, netcontrolintent string, exists bool) (WorkloadIntent, error)
	GetWorkloadIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent string) (WorkloadIntent, error)
	GetWorkloadIntents(project, compositeapp, compositeappversion, dig, netcontrolintent string) ([]WorkloadIntent, error)
	DeleteWorkloadIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent string) error
}

// WorkloadIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkloadIntentClient struct {
	db ClientDbInfo
}

// NewWorkloadIntentClient returns an instance of the WorkloadIntentClient
// which implements the Manager
func NewWorkloadIntentClient() *WorkloadIntentClient {
	return &WorkloadIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "workloadintentmetadata",
		},
	}
}

// CreateWorkloadIntent - create a new WorkloadIntent
func (v *WorkloadIntentClient) CreateWorkloadIntent(wi WorkloadIntent, project, compositeapp, compositeappversion, dig, netcontrolintent string, exists bool) (WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      wi.Metadata.Name,
	}

	//Check if the Network Control Intent exists
	_, err := NewNetControlIntentClient().GetNetControlIntent(netcontrolintent, project, compositeapp, compositeappversion, dig)
	if err != nil {
		return WorkloadIntent{}, pkgerrors.Errorf("Network Control Intent %v does not exist", netcontrolintent)
	}

	//Check if this WorkloadIntent already exists
	_, err = v.GetWorkloadIntent(wi.Metadata.Name, project, compositeapp, compositeappversion, dig, netcontrolintent)
	if err == nil && !exists {
		return WorkloadIntent{}, pkgerrors.New("WorkloadIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, wi)
	if err != nil {
		return WorkloadIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return wi, nil
}

// GetWorkloadIntent returns the WorkloadIntent for corresponding name
func (v *WorkloadIntentClient) GetWorkloadIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent string) (WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return WorkloadIntent{}, pkgerrors.Wrap(err, "Get WorkloadIntent")
	}

	//value is a byte array
	if value != nil {
		wi := WorkloadIntent{}
		err = db.DBconn.Unmarshal(value[0], &wi)
		if err != nil {
			return WorkloadIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return wi, nil
	}

	return WorkloadIntent{}, pkgerrors.New("Error getting WorkloadIntent")
}

// GetWorkloadIntentList returns all of the WorkloadIntent for corresponding name
func (v *WorkloadIntentClient) GetWorkloadIntents(project, compositeapp, compositeappversion, dig, netcontrolintent string) ([]WorkloadIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      "",
	}

	var resp []WorkloadIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []WorkloadIntent{}, pkgerrors.Wrap(err, "Get WorkloadIntents")
	}

	for _, value := range values {
		wi := WorkloadIntent{}
		err = db.DBconn.Unmarshal(value, &wi)
		if err != nil {
			return []WorkloadIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, wi)
	}

	return resp, nil
}

// Delete the  WorkloadIntent from database
func (v *WorkloadIntentClient) DeleteWorkloadIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent string) error {

	//Construct key and tag to select the entry
	key := WorkloadIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete WorkloadIntent Entry;")
	}

	return nil
}
