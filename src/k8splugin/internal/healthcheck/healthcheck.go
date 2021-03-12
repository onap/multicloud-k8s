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

// HealthcheckState holds possible states of Healthcheck instance
type HealthcheckState string

const (
	HcS_UNKNOWN HealthcheckState = "UNKNOWN"
	HcS_STARTED HealthcheckState = "STARTED"
	HcS_SUCCESS HealthcheckState = "SUCCESS"
	HcS_FAILURE HealthcheckState = "FAILURE"
)

// InstanceHCManager interface exposes instance Healthcheck fuctionalities
type InstanceHCManager interface {
	Create(instanceId string) (InstanceHCStatus, error)
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
	InstanceId               string           `json:"instance-id"`
	HealthcheckId            string           `json:"healthcheck-id"`
	Status                   HealthcheckState `json:"status"`
	Info                     string           `json:"info"`
	releasetesting.TestSuite `json:"test-suite"`
}

// InstanceMiniHCStatus holds only healthcheck summary
type InstanceMiniHCStatus struct {
	HealthcheckId string           `json:"healthcheck-id"`
	Status        HealthcheckState `json:"status"`
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

func (ihc InstanceHCClient) Create(instanceId string) (InstanceHCStatus, error) {
	//Generate ID
	id := namegenerator.Generate()

	//Determine Cloud Region and namespace
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Getting instance")
	}

	//Prepare Environment, Request and Release structs
	//TODO In future could derive params from request
	client, err := NewKubeClient(instanceId, instance.Request.CloudRegion)
	if err != nil {
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Preparing KubeClient")
	}
	env := &releasetesting.Environment{
		Namespace:  instance.Namespace,
		KubeClient: client,
		Parallel:   false,
		Stream:     NewStream(),
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
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Creating TestSuite")
	}

	//Save state
	ihcs := InstanceHCStatus{
		TestSuite:     *testSuite,
		HealthcheckId: id,
		InstanceId:    instanceId,
		Status:        HcS_STARTED,
	}
	key := HealthcheckKey{
		InstanceId:    instanceId,
		HealthcheckId: id,
	}
	err = db.DBconn.Create(ihc.storeName, key, ihc.tagInst, ihcs)
	if err != nil {
		return ihcs, pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	// Run HC async
	// If testSuite doesn't fail immediately, we let it continue in background
	errC := make(chan error, 1)
	timeout := make(chan bool, 1)
	// Stream handles updates of testSuite run so we don't need to care
	go func() {
		err := testSuite.Run(env)
		if err != nil {
			log.Error("Error running TestSuite", log.Fields{
				"TestSuite":   testSuite,
				"Environment": env,
				"Error":       err.Error(),
			})
			ihcs.Status = HcS_FAILURE
			ihcs.Info = err.Error()
		} else {
			//TODO handle same way as in GET /hc/$hcid
			ihcs.Status = HcS_UNKNOWN
			ihcs.Info = "TestSuite ended but results were not parsed"
		}
		// Send to channel before db update as it can be slow
		errC <- err
		err = db.DBconn.Update(ihc.storeName, key, ihc.tagInst, ihcs)
		if err != nil {
			log.Error("Error saving Testsuite failure in DB", log.Fields{
				"InstanceHCStatus": ihcs,
				"Error":            err,
			})
		}
	}()
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
	}()
	select {
	case err := <-errC:
		if err == nil {
			return ihcs, nil
		} else {
			return ihcs, err
		}
	case <-timeout:
		return ihcs, nil
	}
	//FIXME Remove
	/*
		if err = testSuite.Run(env); err != nil {
			log.Error("Error running TestSuite", log.Fields{
				"TestSuite":   testSuite,
				"Environment": env,
			})
			return InstanceHCStatus{}, pkgerrors.Wrap(err, "Running TestSuite")
		}
		return ihcs, nil
	*/
}

func (ihc InstanceHCClient) Get(instanceId, healthcheckId string) (InstanceHCStatus, error) {
	return InstanceHCStatus{}, nil
}

func (ihc InstanceHCClient) Delete(instanceId, healthcheckId string) error {
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
			miniStatus = append(miniStatus, InstanceMiniHCStatus{
				HealthcheckId: resp.HealthcheckId,
				Status:        resp.Status,
			})
		}
	}
	ihco.HCSummary = miniStatus

	return ihco, nil
}
