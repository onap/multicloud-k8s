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

package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/resourcestatus"
	kubeclient "github.com/onap/multicloud-k8s/src/rsync/pkg/client"
	connector "github.com/onap/multicloud-k8s/src/rsync/pkg/connector"
	utils "github.com/onap/multicloud-k8s/src/rsync/pkg/internal"
	status "github.com/onap/multicloud-k8s/src/rsync/pkg/status"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CompositeAppContext struct {
	cid   interface{}
	chans []chan bool
	mutex sync.Mutex
}

func getRes(ac appcontext.AppContext, name string, app string, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte
	rh, err := ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		return nil, nil, err
	}
	sh, err := ac.GetLevelHandle(rh, "status")
	if err != nil {
		return nil, nil, err
	}
	resval, err := ac.GetValue(rh)
	if err != nil {
		return nil, sh, err
	}
	if resval != "" {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return nil, sh, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		return nil, sh, pkgerrors.Errorf("Resource value is nil %s", name)
	}
	return byteRes, sh, nil
}

func getSubResApprove(ac appcontext.AppContext, name string, app string, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte
	rh, err := ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		return nil, nil, err
	}
	sh, err := ac.GetLevelHandle(rh, "subresource/approval")
	if err != nil {
		return nil, nil, err
	}
	resval, err := ac.GetValue(sh)
	if err != nil {
		return nil, sh, err
	}
	if resval != "" {
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		return nil, sh, pkgerrors.Errorf("SubResource value is nil %s", name)
	}
	return byteRes, sh, nil
}

func terminateResource(ac appcontext.AppContext, c *kubeclient.Client, name string, app string, cluster string, label string) error {
	res, sh, err := getRes(ac, name, app, cluster)
	if err != nil {
		if sh != nil {
			ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		}
		return err
	}
	if err := c.Delete(res); err != nil {
		ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		logutils.Error("Failed to delete res", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Deleted})
	logutils.Info("Deleted::", logutils.Fields{
		"cluster":  cluster,
		"resource": name,
	})
	return nil
}

func instantiateResource(ac appcontext.AppContext, c *kubeclient.Client, name string, app string, cluster string, label string) error {
	res, sh, err := getRes(ac, name, app, cluster)
	if err != nil {
		if sh != nil {
			ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		}
		return err
	}
	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err = utils.DecodeYAMLData(string(res), unstruct)
	if err != nil {
		return pkgerrors.Wrap(err, "Decode deployment object error")
	}

	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	//labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	labels["emco/deployment-id"] = label
	unstruct.SetLabels(labels)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	//connector.TagPodsIfPresent(unstruct, client.GetInstanceID())
	utils.TagPodsIfPresent(unstruct, label)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		logutils.Error("Failed to MarshalJSON", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	if err := c.Apply(b); err != nil {
		ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		logutils.Error("Failed to apply res", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	ac.UpdateStatusValue(sh, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	logutils.Info("Installed::", logutils.Fields{
		"cluster":  cluster,
		"resource": name,
	})

	// Currently only subresource supported is approval
	subres, _, err := getSubResApprove(ac, name, app, cluster)
	if err == nil {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		logutils.Info("Approval Subresource::", logutils.Fields{
			"cluster":  cluster,
			"resource": result[0],
			"approval": string(subres),
		})
		err = c.Approve(result[0], subres)
		return err
	}
	return nil
}

func updateResourceStatus(ac appcontext.AppContext, resState resourcestatus.ResourceStatus, app string, cluster string, aov map[string][]string) error {

	for _, res := range aov["resorder"] {

		rh, err := ac.GetResourceHandle(app, cluster, res)
		if err != nil {
			return err
		}
		sh, err := ac.GetLevelHandle(rh, "status")
		if err != nil {
			return err
		}

		s, err := ac.GetValue(sh)
		if err != nil {
			return err
		}
		rStatus := resourcestatus.ResourceStatus{}
		js, err := json.Marshal(s)
		if err != nil {
			return err
		}
		err = json.Unmarshal(js, &rStatus)
		if err != nil {
			return err
		}
		// no need to update a status that has reached a 'done' status
		if rStatus.Status == resourcestatus.RsyncStatusEnum.Deleted ||
			rStatus.Status == resourcestatus.RsyncStatusEnum.Applied ||
			rStatus.Status == resourcestatus.RsyncStatusEnum.Failed {
			continue
		}

		err = ac.UpdateStatusValue(sh, resState)
		if err != nil {
			return err
		}
	}

	return nil

}

// return true if all resources have reached a 'done' status - e.g. Applied, Deleted or Failed
func allResourcesDone(ac appcontext.AppContext, app string, cluster string, aov map[string][]string) bool {

	for _, res := range aov["resorder"] {

		rh, err := ac.GetResourceHandle(app, cluster, res)
		if err != nil {
			return false
		}
		sh, err := ac.GetLevelHandle(rh, "status")
		if err != nil {
			return false
		}

		s, err := ac.GetValue(sh)
		if err != nil {
			return false
		}
		rStatus := resourcestatus.ResourceStatus{}
		js, err := json.Marshal(s)
		if err != nil {
			return false
		}
		err = json.Unmarshal(js, &rStatus)
		if err != nil {
			return false
		}
		if rStatus.Status != resourcestatus.RsyncStatusEnum.Deleted &&
			rStatus.Status != resourcestatus.RsyncStatusEnum.Applied &&
			rStatus.Status != resourcestatus.RsyncStatusEnum.Failed {
			return false
		}
	}

	return true

}

// Wait for 2 secs
const waitTime = 2

func waitForClusterReady(instca *CompositeAppContext, ac appcontext.AppContext, c *kubeclient.Client, appname string, cluster string, aov map[string][]string) error {

	forceDone := false
	resStateUpdated := false
	ch := addChan(instca)

	rch := make(chan error, 1)
	checkReachable := func() {
		err := c.IsReachable()
		rch <- err
	}

	go checkReachable()
Loop:
	for {
		select {
		case rerr := <-rch:
			if rerr == nil {
				break Loop
			} else {
				logutils.Info("Cluster is not reachable - keep trying::", logutils.Fields{"cluster": cluster})
			}
		case <-ch:
			statusFailed := resourcestatus.ResourceStatus{
				Status: resourcestatus.RsyncStatusEnum.Failed,
			}
			err := updateResourceStatus(ac, statusFailed, appname, cluster, aov)
			if err != nil {
				deleteChan(instca, ch)
				return err
			}
			forceDone = true
			break Loop
		case <-time.After(waitTime * time.Second):
			// on first timeout - cluster is apparently not reachable, update resources in
			// this group to 'Retrying'
			if !resStateUpdated {
				statusRetrying := resourcestatus.ResourceStatus{
					Status: resourcestatus.RsyncStatusEnum.Retrying,
				}
				err := updateResourceStatus(ac, statusRetrying, appname, cluster, aov)
				if err != nil {
					deleteChan(instca, ch)
					return err
				}
				resStateUpdated = true
			}
			go checkReachable()
			break
		}
	}

	deleteChan(instca, ch)
	if forceDone {
		return pkgerrors.Errorf("Termination of rsync cluster retry: " + cluster)
	}
	return nil
}

// initializeAppContextStatus sets the initial status of every resource appropriately based on the state of the AppContext
func initializeAppContextStatus(ac appcontext.AppContext, acStatus appcontext.AppContextStatus) error {
	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if sh == nil {
		_, err = ac.AddLevelValue(h, "status", acStatus)
	} else {
		err = ac.UpdateValue(sh, acStatus)
	}
	if err != nil {
		return err
	}
	return nil
}

// initializeResourceStatus sets the initial status of every resource appropriately based on the state of the AppContext
func initializeResourceStatus(ac appcontext.AppContext, acStatus appcontext.AppContextStatus) error {
	statusPending := resourcestatus.ResourceStatus{
		Status: resourcestatus.RsyncStatusEnum.Pending,
	}
	statusDeleted := resourcestatus.ResourceStatus{
		Status: resourcestatus.RsyncStatusEnum.Deleted,
	}

	appsOrder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		clusterNames, err := ac.GetClusterNames(app)
		if err != nil {
			return err
		}
		for k := 0; k < len(clusterNames); k++ {
			cluster := clusterNames[k]
			resorder, err := ac.GetResourceInstruction(app, cluster, "order")
			if err != nil {
				return err
			}
			var aov map[string][]string
			json.Unmarshal([]byte(resorder.(string)), &aov)
			for _, res := range aov["resorder"] {
				rh, err := ac.GetResourceHandle(app, cluster, res)
				if err != nil {
					return err
				}
				sh, err := ac.GetLevelHandle(rh, "status")
				if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating {
					if sh == nil {
						_, err = ac.AddLevelValue(rh, "status", statusPending)
					} else {
						err = ac.UpdateStatusValue(sh, statusPending)
					}
					if err != nil {
						return err
					}
				} else if acStatus.Status == appcontext.AppContextStatusEnum.Terminating {
					if sh == nil {
						_, err = ac.AddLevelValue(rh, "status", statusDeleted)
					} else {
						s, err := ac.GetValue(sh)
						if err != nil {
							return err
						}
						rStatus := resourcestatus.ResourceStatus{}
						js, _ := json.Marshal(s)
						json.Unmarshal(js, &rStatus)
						if rStatus.Status == resourcestatus.RsyncStatusEnum.Applied {
							err = ac.UpdateStatusValue(sh, statusPending)
						} else {
							err = ac.UpdateStatusValue(sh, statusDeleted)
						}
						if err != nil {
							return err
						}
					}
				} else {
					return pkgerrors.Errorf("Error intializing AppContext Resource Statuses")
				}
			}
		}
	}
	return nil
}

func addStatusTracker(c *kubeclient.Client, app string, cluster string, label string) error {

	b, err := status.GetStatusCR(label)
	if err != nil {
		logutils.Error("Failed to get status CR for installing", logutils.Fields{
			"error": err,
			"label": label,
		})
		return err
	}
	// TODO: Check reachability?
	if err = c.Apply(b); err != nil {
		logutils.Error("Failed to apply status tracker", logutils.Fields{
			"error":   err,
			"cluster": cluster,
			"app":     app,
			"label":   label,
		})
		return err
	}
	logutils.Info("Status tracker installed::", logutils.Fields{
		"cluster": cluster,
		"app":     app,
		"label":   label,
	})
	return nil
}

func deleteStatusTracker(c *kubeclient.Client, app string, cluster string, label string) error {
	b, err := status.GetStatusCR(label)
	if err != nil {
		logutils.Error("Failed to get status CR for deleting", logutils.Fields{
			"error": err,
			"label": label,
		})
		return err
	}
	if err = c.Delete(b); err != nil {
		logutils.Error("Failed to delete res", logutils.Fields{
			"error": err,
			"app":   app,
			"label": label,
		})
		return err
	}
	logutils.Info("Status tracker deleted::", logutils.Fields{
		"cluster": cluster,
		"app":     app,
		"label":   label,
	})
	return nil
}

func updateEndingAppContextStatus(ac appcontext.AppContext, handle interface{}, failure bool) error {
	sh, err := ac.GetLevelHandle(handle, "status")
	if err != nil {
		return err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating {
		if failure {
			acStatus.Status = appcontext.AppContextStatusEnum.InstantiateFailed
		} else {
			acStatus.Status = appcontext.AppContextStatusEnum.Instantiated
		}
	} else if acStatus.Status == appcontext.AppContextStatusEnum.Terminating {
		if failure {
			acStatus.Status = appcontext.AppContextStatusEnum.TerminateFailed
		} else {
			acStatus.Status = appcontext.AppContextStatusEnum.Terminated
		}
	} else {
		return pkgerrors.Errorf("Invalid AppContextStatus %v", acStatus)
	}

	err = ac.UpdateValue(sh, acStatus)
	if err != nil {
		return err
	}
	return nil
}

func getAppContextStatus(ac appcontext.AppContext) (*appcontext.AppContextStatus, error) {

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return nil, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return nil, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return nil, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return &acStatus, nil

}

type fn func(ac appcontext.AppContext, client *kubeclient.Client, res string, app string, cluster string, label string) error

type statusfn func(client *kubeclient.Client, app string, cluster string, label string) error

func addChan(instca *CompositeAppContext) chan bool {

	instca.mutex.Lock()
	c := make(chan bool)
	instca.chans = append(instca.chans, c)
	instca.mutex.Unlock()

	return c
}

func deleteChan(instca *CompositeAppContext, c chan bool) error {

	var i int
	instca.mutex.Lock()
	for i = 0; i < len(instca.chans); i++ {
		if instca.chans[i] == c {
			break
		}
	}

	if i == len(instca.chans) {
		instca.mutex.Unlock()
		return pkgerrors.Errorf("Given channel was not found:")
	}
	instca.chans[i] = instca.chans[len(instca.chans)-1]
	instca.chans = instca.chans[:len(instca.chans)-1]
	instca.mutex.Unlock()

	return nil
}

func waitForDone(ac appcontext.AppContext) {
	count := 0
	for {
		time.Sleep(1 * time.Second)
		count++
		if count == 60*60 {
			logutils.Info("Wait for done watcher running..", logutils.Fields{})
			count = 0
		}
		acStatus, err := getAppContextStatus(ac)
		if err != nil {
			logutils.Error("Failed to get the app context status", logutils.Fields{
				"error": err,
			})
			return
		}
		if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated ||
			acStatus.Status == appcontext.AppContextStatusEnum.InstantiateFailed {
			return
		}
	}
	return
}

func kickoffRetryWatcher(instca *CompositeAppContext, ac appcontext.AppContext, acStatus appcontext.AppContextStatus, wg *errgroup.Group) {

	wg.Go(func() error {

		var count int

		count = 0
		for {
			time.Sleep(1 * time.Second)
			count++
			if count == 60*60 {
				logutils.Info("Retry watcher running..", logutils.Fields{})
				count = 0
			}

			cStatus, err := getAppContextStatus(ac)
			if err != nil {
				logutils.Error("Failed to get the app context status", logutils.Fields{
					"error": err,
				})
				return err
			}
			flag, err := getAppContextFlag(ac)
			if err != nil {
				logutils.Error("Failed to get the stop flag", logutils.Fields{
					"error": err,
				})
				return err
			} else {
				if flag == true {
					instca.mutex.Lock()
					for i := 0; i < len(instca.chans); i++ {
						instca.chans[i] <- true
						logutils.Info("kickoffRetryWatcher - send an exit message", logutils.Fields{})
					}
					instca.mutex.Unlock()
					break
				}
			}
			if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating {
				if cStatus.Status == appcontext.AppContextStatusEnum.Instantiated ||
					cStatus.Status == appcontext.AppContextStatusEnum.InstantiateFailed {
					break
				}
			} else {
				if cStatus.Status == appcontext.AppContextStatusEnum.Terminated ||
					cStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed {
					break
				}
			}

		}
		return nil
	})

}

func getAppContextFlag(ac appcontext.AppContext) (bool, error) {
	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return false, err
	}
	sh, err := ac.GetLevelHandle(h, "stopflag")
	if sh == nil {
		return false, err
	} else {
		v, err := ac.GetValue(sh)
		if err != nil {
			return false, err
		} else {
			return v.(bool), nil
		}
	}
}

func updateAppContextFlag(cid interface{}, sf bool) error {
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
	}
	hc, err := ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(hc, "stopflag")
	if sh == nil {
		_, err = ac.AddLevelValue(hc, "stopflag", sf)
	} else {
		err = ac.UpdateValue(sh, sf)
	}
	if err != nil {
		return err
	}
	return nil
}

func applyFnComApp(instca *CompositeAppContext, acStatus appcontext.AppContextStatus, f fn, sfn statusfn, breakonError bool) error {
	con := connector.Init(instca.cid)
	//Cleanup
	defer con.RemoveClient()
	ac := appcontext.AppContext{}
	h, err := ac.LoadAppContext(instca.cid)
	if err != nil {
		return err
	}

	// if terminating, wait for all retrying instantiate threads to exit
	if acStatus.Status == appcontext.AppContextStatusEnum.Terminating {
		waitForDone(ac)
		err := updateAppContextFlag(instca.cid, false)
		if err != nil {
			return err
		}
	}

	// initialize appcontext status
	err = initializeAppContextStatus(ac, acStatus)
	if err != nil {
		return err
	}

	// initialize the resource status values before proceeding with the function
	err = initializeResourceStatus(ac, acStatus)
	if err != nil {
		return err
	}

	appsOrder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)
	logutils.Info("appsorder ", logutils.Fields{
		"appsorder": appsOrder,
		"string":    appList,
	})
	id, _ := ac.GetCompositeAppHandle()
	g, _ := errgroup.WithContext(context.Background())
	wg, _ := errgroup.WithContext(context.Background())
	kickoffRetryWatcher(instca, ac, acStatus, wg)
	// Iterate over all the subapps
	for _, app := range appList["apporder"] {
		appName := app
		results := strings.Split(id.(string), "/")
		label := results[2] + "-" + app
		g.Go(func() error {
			clusterNames, err := ac.GetClusterNames(appName)
			if err != nil {
				return err
			}
			// Iterate over all clusters
			for k := 0; k < len(clusterNames); k++ {
				cluster := clusterNames[k]
				err = status.StartClusterWatcher(cluster)
				if err != nil {
					logutils.Error("Error starting Cluster Watcher", logutils.Fields{
						"error":   err,
						"cluster": cluster,
					})
				}
				g.Go(func() error {
					c, err := con.GetClient(cluster)
					if err != nil {
						logutils.Error("Error in creating kubeconfig client", logutils.Fields{
							"error":   err,
							"cluster": cluster,
							"appName": appName,
						})
						return err
					}
					resorder, err := ac.GetResourceInstruction(appName, cluster, "order")
					if err != nil {
						logutils.Error("Resorder error ", logutils.Fields{"error": err})
						return err
					}
					var aov map[string][]string
					json.Unmarshal([]byte(resorder.(string)), &aov)
					// Keep retrying for reachability
					for {
						done := allResourcesDone(ac, appName, cluster, aov)
						if done {
							break
						}

						// Wait for cluster to be reachable
						err := waitForClusterReady(instca, ac, c, appName, cluster, aov)
						if err != nil {
							// TODO: Add error handling
							return err
						}
						reachable := true
						// Handle all resources in order
						for i, res := range aov["resorder"] {
							err = f(ac, c, res, appName, cluster, label)
							if err != nil {
								logutils.Error("Error in resource %s: %v", logutils.Fields{
									"error":    err,
									"cluster":  cluster,
									"resource": res,
								})
								// If failure is due to reachability issues start retrying
								if err = c.IsReachable(); err != nil {
									reachable = false
									break
								}
								if breakonError {
									// handle status tracking before exiting if at least one resource got handled
									if i > 0 {
										serr := sfn(c, appName, cluster, label)
										if serr != nil {
											logutils.Warn("Error handling status tracker", logutils.Fields{"error": serr})
										}
									}
									return err
								}
							}
						}
						// Check if the break from loop due to reachabilty issues
						if reachable != false {
							serr := sfn(c, appName, cluster, label)
							if serr != nil {
								logutils.Warn("Error handling status tracker", logutils.Fields{"error": serr})
							}
							// Done processing cluster without errors
							return nil
						}
					}
					return nil
				})
			}
			return nil
		})
	}
	// Wait for all subtasks to complete
	if err := g.Wait(); err != nil {
		uperr := updateEndingAppContextStatus(ac, h, true)
		if uperr != nil {
			logutils.Error("Encountered error updating AppContext to Failed status", logutils.Fields{"error": uperr})
		}
		logutils.Error("Encountered error", logutils.Fields{
			"error": err,
		})
		return err
	}
	err = updateEndingAppContextStatus(ac, h, false)
	if err != nil {
		logutils.Error("Encountered error updating AppContext status", logutils.Fields{"error": err})
		return err
	}
	if err := wg.Wait(); err != nil {
		logutils.Error("Encountered error in watcher thread", logutils.Fields{"error": err})
		return err
	}
	return nil
}

// InstantiateComApp Instantiate Apps in Composite App
func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	instca.cid = cid
	instca.chans = []chan bool{}
	instca.mutex = sync.Mutex{}
	err := updateAppContextFlag(cid, false)
	if err != nil {
		logutils.Error("Encountered error updating AppContext flag", logutils.Fields{"error": err})
		return err
	}
	go applyFnComApp(instca, appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiating},
		instantiateResource, addStatusTracker, true)
	return nil
}

// TerminateComApp Terminates Apps in Composite App
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	instca.cid = cid
	instca.chans = []chan bool{}
	instca.mutex = sync.Mutex{}
	err := updateAppContextFlag(cid, true)
	if err != nil {
		logutils.Error("Encountered error updating AppContext flag", logutils.Fields{"error": err})
		return err
	}
	go applyFnComApp(instca, appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Terminating},
		terminateResource, deleteStatusTracker, false)
	return nil
}
