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

package quota

import (
        "encoding/json"

        "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

        pkgerrors "github.com/pkg/errors"
)

// Quota contains the parameters needed for a Quota 
type Quota struct {
        MetaData            MetaDataList        `json:"metadata"`
        Specification       Spec                `json:"spec"`
}


// MetaData contains the parameters needed for metadata
type MetaDataList struct {
        QuotaName           string  `json:"name"`
        Description         string  `json:"description"`
}

// Spec contains the parameters needed for spec
type Spec struct {
        LimitsCPU                   string      `json:"limits.cpu"`
        LimitsMemory                string      `json:"limits.memory"`
        RequestsCPU                 string      `json:"requests.cpu"`
        RequestsMemory              string      `json:"requests.memory"`
        RequestsStorage             string      `json:"requests.storage"`
        LimitsEphemeralStorage      string      `json:"limits.ephemeral.storage"`
        PersistentVolumeClaims      string      `json:"persistentvolumeclaims"`
        Pods                        string      `json:"pods"`
        ConfigMaps                  string      `json:"configmaps"`
        ReplicationControllers      string      `json:"replicationcontrollers"`
        ResourceQuotas              string      `json:"resourcequotas"`
        Services                    string      `json:"services"`
        ServicesLoadBalancers       string      `json:"services.loadbalancers"`
        ServicesNodePorts           string      `json:"services.nodeports"`
        Secrets                     string      `json:"secrets"`
        CountReplicationControllers string      `json:"count/replicationcontrollers"`
        CountDeploymentsApps        string      `json:"count/deployments.apps"`
        CountReplicasetsApps        string      `json:"count/replicasets.apps"`
        CountStatefulSets           string      `json:"count/statefulsets.apps"`
        CountJobsBatch              string      `json:"count/jobs.batch"`
        CountCronJobsBatch          string      `json:"count/cronjobs.batch"`
        CountDeploymentsExtensions  string      `json:"count/deployments.extensions"`

}

// QuotaKey is the key structure that is used in the database
type QuotaKey struct {
        QuotaName    string `json:"name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure
func (dk QuotaKey) String() string {
        out, err := json.Marshal(dk)
        if err != nil {
                return ""
        }

        return string(out)
}

// QuotaManager is an interface that exposes the connection
// functionality
type QuotaManager interface {
        Create(c Quota) (Quota, error)
        Get(name string) (Quota, error)
        Delete(name string) error
        Update(name string, c Quota) (Quota, error)
}

// QuotaClient implements the QuotaManager
// It will also be used to maintain some localized state
type QuotaClient struct {
        storeName   string
        tagMeta     string
}

// QuotaClient returns an instance of the QuotaClient
// which implements the QuotaManager
func NewQuotaClient() *QuotaClient {
        return &QuotaClient{
               storeName:   "quota",
               tagMeta:     "metadata",
        }
}

// Create entry for the quota resource in the database
func (v *QuotaClient) Create(c Quota) (Quota, error) {

        //Construct key consisting of name
        key := QuotaKey{QuotaName: c.MetaData.QuotaName}

        //Check if this Quota already exists
        _, err := v.Get(c.MetaData.QuotaName)
        if err == nil {
                return Quota{}, pkgerrors.New("Quota already exists")
        }

        err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
        if err != nil {
                return Quota{}, pkgerrors.Wrap(err, "Creating DB Entry")
        }

        return c, nil
}

// Get returns Quota for correspondin name
func (v *QuotaClient) Get(name string) (Quota, error) {

        //Construct the composite key to select the entry
        key := QuotaKey{QuotaName: name}
        value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
        if err != nil {
                return Quota{}, pkgerrors.Wrap(err, "Quota")
        }

        //value is a byte array
        if value != nil {
                q := Quota{}
                err = db.DBconn.Unmarshal(value, &q)
                if err != nil {
                        return Quota{}, pkgerrors.Wrap(err, "Unmarshaling value")
                }
                return q, nil
        }

        return Quota{}, pkgerrors.New("Error getting Quota")
}

// Delete the Quota entry from database
func (v *QuotaClient) Delete(name string) error {
        //Construct the composite key to select the entry
        key := QuotaKey{QuotaName: name}
        err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
        if err != nil {
                return pkgerrors.Wrap(err, "Quota")
        }
        return nil
}

// Update an entry for the Quota in the database
func (v *QuotaClient) Update(name string, c Quota) (Quota, error) {

    key := QuotaKey{
        QuotaName: name,
    }
    //Check if this Quota exists
    _, err := v.Get(name)
    if err != nil {
        return Quota{}, pkgerrors.New("Update Error - Quota doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return Quota{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}

