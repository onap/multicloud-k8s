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
	HcS_RUNNING HealthcheckState = "RUNNING"
	HcS_SUCCESS HealthcheckState = "SUCCESS"
	HcS_FAILURE HealthcheckState = "FAILURE"
)

// InstanceHCManager interface exposes instance Healthcheck fuctionalities
type InstanceHCManager interface {
	Create(instanceId string) (InstanceHCStatus, error)
	Get(instanceId, healthcheckId string) (InstanceHCStatus, error)
	List(instanceId string) ([]InstanceHCStatus, error)
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
	releasetesting.TestSuite
	Id     string
	Status HealthcheckState
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
	}
	release := protorelease.Release{
		Name:  instance.ReleaseName,
		Hooks: instance.Hooks,
	}

	//Run HC
	testSuite, err := releasetesting.NewTestSuite(&release)
	if err != nil {
		log.Error("Error creating TestSuite", log.Fields{
			"Release": release,
		})
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Creating TestSuite")
	}
	if err = testSuite.Run(env); err != nil {
		log.Error("Error running TestSuite", log.Fields{
			"TestSuite":   testSuite,
			"Environment": env,
		})
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Running TestSuite")
	}

	//Save state
	ihcs := InstanceHCStatus{
		TestSuite: *testSuite,
		Id:        id,
		Status:    HcS_STARTED,
	}
	key := HealthcheckKey{
		InstanceId:    instance.ID,
		HealthcheckId: id,
	}
	err = db.DBconn.Create(ihc.storeName, key, ihc.tagInst, ihcs)
	if err != nil {
		return ihcs, pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}

	return ihcs, nil
}

func (ihc InstanceHCClient) Get(instanceId, healthcheckId string) (InstanceHCStatus, error) {
	return InstanceHCStatus{}, nil
}

func (ihc InstanceHCClient) Delete(instanceId, healthcheckId string) error {
	return nil
}

func (ihc InstanceHCClient) List(instanceId string) ([]InstanceHCStatus, error) {
	return make([]InstanceHCStatus, 0), nil
}
