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
	"k8s.io/helm/pkg/releasetesting"
	//"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	pkgerrors "github.com/pkg/errors"
)

// InstanceHCManager interface exposes instance Healthcheck fuctionalities
type InstanceHCManager interface {
	Create(instanceId string) (InstanceHCStatus, error)
	Get(instanceId, healthcheckId string) (InstanceHCStatus, error)
	List(instanceId string) ([]InstanceHCStatus, error)
	Delete(instanceId, healthcheckId string) error
}

// InstanceHCClient implements InstanceHCManager
type InstanceHCClient struct {
}

// InstanceHCStatus holds healthcheck status
type InstanceHCStatus struct {
	releasetesting.TestSuite
	Id     string
	Status string
}

func NewHCClient() *InstanceHCClient {
	return &InstanceHCClient{}
}

func ensureInstanceExists(instanceId string) error {
	client := app.NewInstanceClient()
	_, err := client.Get(instanceId)
	if err != nil {
		return pkgerrors.New("Instance not found running")
	}
	return nil
}

func (ihc InstanceHCClient) Create(instanceId string) (InstanceHCStatus, error) {
	if err := ensureInstanceExists(instanceId); err != nil {
		return InstanceHCStatus{}, pkgerrors.Wrap(err, "Checking Instance")
	}
	//Determine Cloud Region and namespace

	//Determine merged Values

	//Prepare "Environment"

	//Derive test resources

	//Run HC

	//Return
	return InstanceHCStatus{}, nil
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
