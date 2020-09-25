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
	pkgerrors "github.com/pkg/errors"
)

// Quota contains the parameters needed for a Quota
type Quota struct {
	MetaData QMetaDataList `json:"metadata"`
	// Specification QSpec         `json:"spec"`
	Specification map[string]string `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type QMetaDataList struct {
	QuotaName   string `json:"name"`
	Description string `json:"description"`
}

// TODO: use QSpec fields to validate quota keys
// Spec contains the parameters needed for spec
type QSpec struct {
	LimitsCPU                   string `json:"limits.cpu"`
	LimitsMemory                string `json:"limits.memory"`
	RequestsCPU                 string `json:"requests.cpu"`
	RequestsMemory              string `json:"requests.memory"`
	RequestsStorage             string `json:"requests.storage"`
	LimitsEphemeralStorage      string `json:"limits.ephemeral.storage"`
	PersistentVolumeClaims      string `json:"persistentvolumeclaims"`
	Pods                        string `json:"pods"`
	ConfigMaps                  string `json:"configmaps"`
	ReplicationControllers      string `json:"replicationcontrollers"`
	ResourceQuotas              string `json:"resourcequotas"`
	Services                    string `json:"services"`
	ServicesLoadBalancers       string `json:"services.loadbalancers"`
	ServicesNodePorts           string `json:"services.nodeports"`
	Secrets                     string `json:"secrets"`
	CountReplicationControllers string `json:"count/replicationcontrollers"`
	CountDeploymentsApps        string `json:"count/deployments.apps"`
	CountReplicasetsApps        string `json:"count/replicasets.apps"`
	CountStatefulSets           string `json:"count/statefulsets.apps"`
	CountJobsBatch              string `json:"count/jobs.batch"`
	CountCronJobsBatch          string `json:"count/cronjobs.batch"`
	CountDeploymentsExtensions  string `json:"count/deployments.extensions"`
}

// QuotaKey is the key structure that is used in the database
type QuotaKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	QuotaName        string `json:"qname"`
}

// QuotaManager is an interface that exposes the connection
// functionality
type QuotaManager interface {
	CreateQuota(project, logicalCloud string, c Quota) (Quota, error)
	GetQuota(project, logicalCloud, name string) (Quota, error)
	GetAllQuotas(project, logicalCloud string) ([]Quota, error)
	DeleteQuota(project, logicalCloud, name string) error
	UpdateQuota(project, logicalCloud, name string, c Quota) (Quota, error)
}

// QuotaClient implements the QuotaManager
// It will also be used to maintain some localized state
type QuotaClient struct {
	storeName string
	tagMeta   string
	util      Utility
}

// QuotaClient returns an instance of the QuotaClient
// which implements the QuotaManager
func NewQuotaClient() *QuotaClient {
	service := DBService{}
	return &QuotaClient{
		storeName: "orchestrator",
		tagMeta:   "quota",
		util:      service,
	}
}

// Create entry for the quota resource in the database
func (v *QuotaClient) CreateQuota(project, logicalCloud string, c Quota) (Quota, error) {

	//Construct key consisting of name
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        c.MetaData.QuotaName,
	}

	//Check if project exists
	err := v.util.CheckProject(project)
	if err != nil {
		return Quota{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = v.util.CheckLogicalCloud(project, logicalCloud)
	if err != nil {
		return Quota{}, pkgerrors.New("Unable to find the logical cloud")
	}
	//Check if this Quota already exists
	_, err = v.GetQuota(project, logicalCloud, c.MetaData.QuotaName)
	if err == nil {
		return Quota{}, pkgerrors.New("Quota already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Quota for corresponding quota name
func (v *QuotaClient) GetQuota(project, logicalCloud, quotaName string) (Quota, error) {

	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Quota")
	}

	//value is a byte array
	if value != nil {
		q := Quota{}
		err = v.util.DBUnmarshal(value[0], &q)
		if err != nil {
			return Quota{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return q, nil
	}

	return Quota{}, pkgerrors.New("Cluster Quota does not exist")
}

// GetAll returns all cluster quotas in the logical cloud
func (v *QuotaClient) GetAllQuotas(project, logicalCloud string) ([]Quota, error) {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        "",
	}
	var resp []Quota
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Quota{}, pkgerrors.Wrap(err, "Get All Quotas")
	}

	for _, value := range values {
		q := Quota{}
		err = v.util.DBUnmarshal(value, &q)
		if err != nil {
			return []Quota{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		resp = append(resp, q)
	}

	return resp, nil
}

// Delete the Quota entry from database
func (v *QuotaClient) DeleteQuota(project, logicalCloud, quotaName string) error {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	err := v.util.DBRemove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Quota")
	}
	return nil
}

// Update an entry for the Quota in the database
func (v *QuotaClient) UpdateQuota(project, logicalCloud, quotaName string, c Quota) (Quota, error) {

	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	//Check quota URL name against the quota json name
	if c.MetaData.QuotaName != quotaName {
		return Quota{}, pkgerrors.New("Update Error - Quota name mismatch")
	}
	//Check if this Quota exists
	_, err := v.GetQuota(project, logicalCloud, quotaName)
	if err != nil {
		return Quota{}, pkgerrors.New("Cluster Quota does not exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
