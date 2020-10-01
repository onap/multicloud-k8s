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
	DigName             string `json:"deploymentintentgroup"`
}

// Manager is an interface exposing the NetControlIntent functionality
type NetControlIntentManager interface {
	CreateNetControlIntent(nci NetControlIntent, project, compositeapp, compositeappversion, dig string, exists bool) (NetControlIntent, error)
	GetNetControlIntent(name, project, compositeapp, compositeappversion, dig string) (NetControlIntent, error)
	GetNetControlIntents(project, compositeapp, compositeappversion, dig string) ([]NetControlIntent, error)
	DeleteNetControlIntent(name, project, compositeapp, compositeappversion, dig string) error
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
func (v *NetControlIntentClient) CreateNetControlIntent(nci NetControlIntent, project, compositeapp, compositeappversion, dig string, exists bool) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    nci.Metadata.Name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	//Check if this NetControlIntent already exists
	_, err := v.GetNetControlIntent(nci.Metadata.Name, project, compositeapp, compositeappversion, dig)
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
func (v *NetControlIntentClient) GetNetControlIntent(name, project, compositeapp, compositeappversion, dig string) (NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
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
func (v *NetControlIntentClient) GetNetControlIntents(project, compositeapp, compositeappversion, dig string) ([]NetControlIntent, error) {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    "",
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
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
func (v *NetControlIntentClient) DeleteNetControlIntent(name, project, compositeapp, compositeappversion, dig string) error {

	//Construct key and tag to select the entry
	key := NetControlIntentKey{
		NetControlIntent:    name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete NetControlIntent Entry;")
	}

	return nil
}
