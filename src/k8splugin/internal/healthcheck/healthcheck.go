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
	"sync"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"

	pkgerrors "github.com/pkg/errors"
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
	InstanceId    string            `json:"instance-id"`
	HealthcheckId string            `json:"healthcheck-id"`
	Status        release.HookPhase `json:"status"`
	TestSuite     *TestSuite        `json:"test-suite"` //TODO could be merged with current struct
}

// TestSuite is a structure to hold compatibility with pre helm3 output
type TestSuite struct {
	StartedAt   time.Time
	CompletedAt time.Time
	Results     []*HookStatus
}

// InstanceMiniHCStatus holds only healthcheck summary
type InstanceMiniHCStatus struct {
	HealthcheckId string            `json:"healthcheck-id"`
	Status        release.HookPhase `json:"status"`
}

// InstanceHCOverview holds Healthcheck-related data
type InstanceHCOverview struct {
	InstanceId string                 `json:"instance-id"`
	HCSummary  []InstanceMiniHCStatus `json:"healthcheck-summary"`
	Hooks      []*helm.Hook           `json:"hooks"`
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
	//TODO Handle hook delete policies

	//Generate ID
	id := namegenerator.Generate()

	//Determine Cloud Region and namespace
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return InstanceMiniHCStatus{}, pkgerrors.Wrap(err, "Getting instance")
	}

	k8sClient := app.KubernetesClient{}
	err = k8sClient.Init(instance.Request.CloudRegion, instanceId)
	if err != nil {
		return InstanceMiniHCStatus{}, pkgerrors.Wrap(err, "Preparing KubeClient")
	}
	key := HealthcheckKey{
		InstanceId:    instanceId,
		HealthcheckId: id,
	}

	//Filter out only relevant hooks
	hooks := hookPairs{}
	for _, hook := range instance.Hooks {
		for _, hookEvent := range hook.Hook.Events {
			if hookEvent == release.HookTest { //Helm3 no longer supports test-failure
				hooks = append(hooks, hookPair{
					Definition: hook,
					Status: &HookStatus{
						Name: hook.Hook.Name,
					},
				})
			}
		}
	}

	//Save state
	testSuite := TestSuite{
		StartedAt: time.Now(),
		Results:   hooks.statuses(),
	}
	ihcs := InstanceHCStatus{
		InstanceId:    instanceId,
		HealthcheckId: id,
		Status:        release.HookPhaseRunning,
		TestSuite:     &testSuite,
	}

	for _, h := range hooks {
		h.Status.StartedAt = time.Now()
		kr, err := k8sClient.CreateKind(h.Definition.KRT, instance.Namespace)
		if err != nil {
			// Set status fields
			h.Status.Status = release.HookPhaseFailed
			h.Status.CompletedAt = time.Now()
			testSuite.CompletedAt = time.Now()
			ihcs.Status = release.HookPhaseFailed
			retErr := "Starting hook " + h.Status.Name

			// Dump to DB
			err = db.DBconn.Create(ihc.storeName, key, ihc.tagInst, ihcs)
			if err != nil {
				retErr = retErr + " and couldn't save to DB"
			}

			return instanceMiniHCStatusFromStatus(ihcs),
				pkgerrors.Wrap(err, retErr)
		} else {
			h.Status.Status = release.HookPhaseRunning
			h.Status.KR = kr
		}
	}
	err = db.DBconn.Create(ihc.storeName, key, ihc.tagInst, ihcs)
	if err != nil {
		return instanceMiniHCStatusFromStatus(ihcs),
			pkgerrors.Wrap(err, "Creating Instance DB Entry")
	}
	log.Info("Successfully initialized Healthcheck resources", log.Fields{
		"InstanceId":    instanceId,
		"HealthcheckId": id,
	})
	go func() {
		var wg sync.WaitGroup
		update := make(chan bool) //True - hook finished, False - all hooks finished
		for _, h := range hooks {
			wg.Add(1)
			go func(hookStatus *HookStatus) {
				//TODO Handle errors here better in future, for now it's ok
				hookStatus.Status, _ = getHookState(*hookStatus, k8sClient, instance.Namespace)
				hookStatus.CompletedAt = time.Now()
				log.Info("Hook finished", log.Fields{
					"HealthcheckId": id,
					"InstanceId":    instanceId,
					"Hook":          hookStatus.Name,
					"Status":        hookStatus.Status,
				})
				update <- true
				wg.Done()
				return
			}(h.Status)
		}
		go func() {
			wg.Wait()
			log.Info("All hooks finished", log.Fields{
				"HealthcheckId": id,
				"InstanceId":    instanceId,
			})
			update <- false
			return
		}()
		for {
			select {
			case b := <-update:
				log.Info("Healthcheck update", log.Fields{
					"HealthcheckId": id,
					"InstanceId":    instanceId,
					"Reason":        map[bool]string{true: "Hook finished", false: "All hooks finished"}[b],
				})
				if b { //Some hook finished - need to update DB
					err = db.DBconn.Update(ihc.storeName, key, ihc.tagInst, ihcs)
					if err != nil {
						log.Error("Couldn't update database", log.Fields{
							"Store":   ihc.storeName,
							"Key":     key,
							"Payload": ihcs,
						})
					}
				} else { //All hooks finished - need to terminate goroutine
					testSuite.CompletedAt = time.Now()
					//If everything's fine, final result is OK
					finalResult := release.HookPhaseSucceeded
					//If at least single hook wasn't determined - it's Unknown
					for _, h := range hooks {
						if h.Status.Status == release.HookPhaseUnknown {
							finalResult = release.HookPhaseUnknown
							break
						}
					}
					//Unless at least one hook failed, then we've failed
					for _, h := range hooks {
						if h.Status.Status == release.HookPhaseFailed {
							finalResult = release.HookPhaseFailed
							break
						}
					}
					ihcs.Status = finalResult
					err = db.DBconn.Update(ihc.storeName, key, ihc.tagInst, ihcs)
					if err != nil {
						log.Error("Couldn't update database", log.Fields{
							"Store":   ihc.storeName,
							"Key":     key,
							"Payload": ihcs,
						})
					}
					return
				}
			}
		}
	}()
	return instanceMiniHCStatusFromStatus(ihcs), nil
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
	return resp, nil
}

func (ihc InstanceHCClient) Delete(instanceId, healthcheckId string) error {
	key := HealthcheckKey{instanceId, healthcheckId}
	v := app.NewInstanceClient()
	instance, err := v.Get(instanceId)
	if err != nil {
		return pkgerrors.Wrap(err, "Getting instance")
	}
	ihcs, err := ihc.Get(instanceId, healthcheckId)
	if err != nil {
		return pkgerrors.Wrap(err, "Error querying Healthcheck status")
	}
	k8sClient := app.KubernetesClient{}
	err = k8sClient.Init(instance.Request.CloudRegion, instanceId)
	if err != nil {
		return pkgerrors.Wrap(err, "Preparing KubeClient")
	}
	cumulatedErr := ""
	for _, hook := range ihcs.TestSuite.Results {
		err = k8sClient.DeleteKind(hook.KR, instance.Namespace)
		//FIXME handle "missing resource" error as not error - hook may be already deleted
		if err != nil {
			cumulatedErr += err.Error() + "\n"
		}
	}
	if cumulatedErr != "" {
		return pkgerrors.New("Removing hooks failed with errors:\n" + cumulatedErr)
	}
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
