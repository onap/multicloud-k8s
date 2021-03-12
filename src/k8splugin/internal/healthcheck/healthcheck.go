/*
Copyright © 2021 Samsung Electronics
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package healthcheck

import (
	"encoding/json"
	"strings"
	"time"

	protorelease "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/releasetesting"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"

	pkgerrors "github.com/pkg/errors"
)

var (
	HcS_UNKNOWN string = protorelease.TestRun_Status_name[int32(protorelease.TestRun_UNKNOWN)]
	HcS_RUNNING string = protorelease.TestRun_Status_name[int32(protorelease.TestRun_RUNNING)]
	HcS_SUCCESS string = protorelease.TestRun_Status_name[int32(protorelease.TestRun_SUCCESS)]
	HcS_FAILURE string = protorelease.TestRun_Status_name[int32(protorelease.TestRun_FAILURE)]
)

// InstanceHCManager interface exposes instance Healthcheck fuctionalities
type InstanceHCManager interface {
	Create(instanceId string) (InstanceMiniHCStatus, error)
	Get(instanceId, healthcheckId string) (InstanceHCStatus, error)
	List(instanceId string) (InstanceHCOverview, error)
	Delete(instanceId, healthcheckId string) error
}

// HealthcheckKey is used as the primary key in the db
type HealthcheckKey struct {
	InstanceId    string `json:"instance-id"`
	HealthcheckId string `json:"healthcheck-id"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk HealthcheckKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// InstanceHCClient implements InstanceHCManager
type InstanceHCClient struct {
	storeName string
	tagInst   string
}

// InstanceHCStatus holds healthcheck status
type InstanceHCStatus struct {
	InstanceId               string `json:"instance-id"`
	HealthcheckId            string `json:"healthcheck-id"`
	Status                   string `json:"status"`
	Info                     string `json:"info"`
	releasetesting.TestSuite `json:"test-suite"`
}

// InstanceMiniHCStatus holds only healthcheck summary
type InstanceMiniHCStatus struct {
	HealthcheckId string `json:"healthcheck-id"`
	Status        string `json:"status"`
}

// InstanceHCOverview holds Healthcheck-related data
type InstanceHCOverview struct {
	InstanceId string                 `json:"instance-id"`
	HCSummary  []InstanceMiniHCStatus `json:"healthcheck-summary"`
	Hooks      []*protorelease.Hook   `json:"hooks"`
}

func NewHCClient() *InstanceHCClient {
	return &InstanceHCClient{
		storeName: "rbdef",
		tagInst:   "instanceHC",
	}
}

func instanceMiniHCStatusFromStatus(ihcs InstanceHCStatus) InstanceMiniHCStatus {
	return InstanceMiniHCStatus{ihcs.HealthcheckId, ihcs.Status}
}

func (ihc InstanceHCClient) Create(instanceId string) (InstanceMiniHCStatus, error) {
	//FIXME switch to InstanceMiniHCStatus
	//Generate ID
	id := namegenerator.Generate()

	//Determine Cloud Region and namespace
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return InstanceMiniHCStatus{}, pkgerrors.Wrap(err, "Getting instance")
	}

	//Prepare Environment, Request and Release structs
	//TODO In future could derive params from request
	client, err := NewKubeClient(instanceId, instance.Request.CloudRegion)
	if err != nil {
		return InstanceMiniHCStatus{}, pkgerrors.Wrap(err, "Preparing KubeClient")
	}
	key := HealthcheckKey{
		InstanceId:    instanceId,
		HealthcheckId: id,
	}
	env := &releasetesting.Environment{
		Namespace:  instance.Namespace,
		KubeClient: client,
		Parallel:   false,
		Stream:     NewStream(key, ihc.storeName, ihc.tagInst),
	}
	release := protorelease.Release{
		Name:  instance.ReleaseName,
		Hooks: instance.Hooks,
	}

	//Define HC
	testSuite, err := releasetesting.NewTestSuite(&release)
	if err != nil {
		log.Error("Error creating TestSuite", log.Fields{
			"Release": release,
		})
		return InstanceMiniHCStatus{}, pkgerrors.Wrap(err, "Creating TestSuite")
	}

	//Save state
	ihcs := InstanceHCStatus{
		TestSuite:     *testSuite,
		HealthcheckId: id,
		InstanceId:    instanceId,
		Status:        HcS_RUNNING,
	}
	err = db.DBconn.Create(ihc.storeName, key, ihc.tagInst, ihcs)
	if err != nil {
		return instanceMiniHCStatusFromStatus(ihcs),
			pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	// Run HC async
	// If testSuite doesn't fail immediately, we let it continue in background
	errC := make(chan error, 1)
	timeout := make(chan bool, 1)
	// Stream handles updates of testSuite run so we don't need to care
	var RunAsync func() = func() {
		err := ihcs.TestSuite.Run(env)
		if err != nil {
			log.Error("Error running TestSuite", log.Fields{
				"TestSuite":   ihcs.TestSuite,
				"Environment": env,
				"Error":       err.Error(),
			})
			ihcs.Status = HcS_FAILURE
			ihcs.Info = err.Error()
		} else {
			//TODO handle same way as in GET /hc/$hcid
		}
		// Send to channel before db update as it can be slow
		errC <- err
		// Download latest Status/Info
		resp, err := ihc.Get(ihcs.InstanceId, ihcs.HealthcheckId)
		if err != nil {
			log.Error("Error querying Healthcheck status", log.Fields{"error": err})
			return
		}
		ihcs.Status = resp.Status
		ihcs.Info = resp.Info
		// Update DB
		err = db.DBconn.Update(ihc.storeName, key, ihc.tagInst, ihcs)
		if err != nil {
			log.Error("Error saving Testsuite failure in DB", log.Fields{
				"InstanceHCStatus": ihcs,
				"Error":            err,
			})
		}
	}
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
	}()
	go RunAsync()
	select {
	case err := <-errC:
		if err == nil {
			return instanceMiniHCStatusFromStatus(ihcs), nil
		} else {
			return instanceMiniHCStatusFromStatus(ihcs), err
		}
	case <-timeout:
		return instanceMiniHCStatusFromStatus(ihcs), nil
	}
}

func (ihc InstanceHCClient) Get(instanceId, healthcheckId string) (InstanceHCStatus, error) {
	key := HealthcheckKey{
		InstanceId:    instanceId,
		HealthcheckId: healthcheckId,
	}
	DBResp, err := db.DBconn.Read(ihc.storeName, key, ihc.tagInst)
	if err != nil || DBResp == nil {
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Error retrieving Healthcheck data")
	}

	resp := InstanceHCStatus{}
	err = db.DBconn.Unmarshal(DBResp, &resp)
	if err != nil {
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Unmarshaling Instance Value")
	}
	//TODO: Implement parsing TestSuite's TestRuns to update Status
	return resp, nil

}

func (ihc InstanceHCClient) Delete(instanceId, healthcheckId string) error {
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return pkgerrors.Wrap(err, "Getting instance")
	}
	client, err := NewKubeClient(instanceId, instance.Request.CloudRegion)
	if err != nil {
		return pkgerrors.Wrap(err, "Preparing KubeClient")
	}
	env := &releasetesting.Environment{
		Namespace:  instance.Namespace,
		KubeClient: client,
	}
	ihcs, err := ihc.Get(instanceId, healthcheckId)
	if err != nil {
		return pkgerrors.Wrap(err, "Error querying Healthcheck status")
	}
	env.DeleteTestPods(ihcs.TestSuite.TestManifests)
	key := HealthcheckKey{instanceId, healthcheckId}
	err = db.DBconn.Delete(ihc.storeName, key, ihc.tagInst)
	if err != nil {
		return pkgerrors.Wrap(err, "Removing Healthcheck in DB")
	}
	return nil
}

func (ihc InstanceHCClient) List(instanceId string) (InstanceHCOverview, error) {
	ihco := InstanceHCOverview{
		InstanceId: instanceId,
	}

	// Retrieve info about available hooks
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return ihco, pkgerrors.Wrap(err, "Getting Instance data")
	}
	ihco.Hooks = instance.Hooks

	// Retrieve info about running/completed healthchecks
	dbResp, err := db.DBconn.ReadAll(ihc.storeName, ihc.tagInst)
	if err != nil {
		if !strings.Contains(err.Error(), "Did not find any objects with tag") {
			return ihco, pkgerrors.Wrap(err, "Getting Healthcheck data")
		}
	}
	miniStatus := make([]InstanceMiniHCStatus, 0)
	for key, value := range dbResp {
		//value is a byte array
		if value != nil {
			resp := InstanceHCStatus{}
			err = db.DBconn.Unmarshal(value, &resp)
			if err != nil {
				log.Error("Error unmarshaling Instance HC data", log.Fields{
					"error": err.Error(),
					"key":   key})
			}
			//Filter-out healthchecks from other instances
			if instanceId != resp.InstanceId {
				continue
			}
			miniStatus = append(miniStatus, instanceMiniHCStatusFromStatus(resp))
		}
	}
	ihco.HCSummary = miniStatus

	return ihco, nil
}
