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
	cid interface{}
}

var chans = []chan bool{}
var mutex = &sync.Mutex{}

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
	return nil
}

// Wait for 2 secs
const waitTime = 2

func waitForClusterReady(c *kubeclient.Client, cluster string) error {
	for {
		if err := c.IsReachable(); err != nil {
			// TODO: Add more realistic error checking
			// TODO: Add Incremental wait logic here
			time.Sleep(waitTime * time.Second)
		} else {
			break
		}
	}
	logutils.Info("Cluster is reachable::", logutils.Fields{
		"cluster": cluster,
	})
	return nil
}

// initializeResourceStatus sets the initial status of every resource appropriately based on the state of the AppContext
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

	if failure {
		acStatus.Status = appcontext.AppContextStatusEnum.Failed
	} else if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating {
		acStatus.Status = appcontext.AppContextStatusEnum.Instantiated
	} else if acStatus.Status == appcontext.AppContextStatusEnum.Terminating {
		acStatus.Status = appcontext.AppContextStatusEnum.Terminated
	} else {
		return pkgerrors.Errorf("Invalid AppContextStatus %v", acStatus)
	}

	err = ac.UpdateValue(sh, acStatus)
	if err != nil {
		return err
	}
	return nil
}

type fn func(ac appcontext.AppContext, client *kubeclient.Client, res string, app string, cluster string, label string) error

type statusfn func(client *kubeclient.Client, app string, cluster string, label string) error

func addChan() chan bool {

	mutex.Lock()
	c := make(chan bool)
	chans = append(chans, c)
	mutex.Unlock()

	return c
}

func deleteChan(c chan bool) error {

	var i int
	mutex.Lock()
	for i =0; i< len(chans); i++ {
		if chans[i] == c {
			break
		}
	}

	if i == len(chans) {
		mutex.Unlock()
		return pkgerrors.Errorf("Given channel was not found:")
	}
	chans[i] = chans[len(chans)-1]
	chans = chans[:len(chans)-1]
	mutex.Unlock()

	return nil
}


func kickoffRetryWatcher(g *errgroup.Group) {

	g.Go(func() error {

		var stop string
		var count int

		count = 0
		for {
			time.Sleep(1 * time.Second)
			count++
			if ( count == 60*60) {
				logutils.Info("Retry watcher running..",logutils.Fields{})
				count = 0
			}
			//Read the signal value from the terminate path
			if ( stop == "stop retrying" ) {

				mutex.Lock()
				for i :=0; i< len(chans); i++ {
					chans[i] <- true
				}
				mutex.Unlock()
				break
			}

		}
		return nil
	})


}

func applyFnComApp(cid interface{}, acStatus appcontext.AppContextStatus, f fn, sfn statusfn, breakonError bool) error {
	con := connector.Init(cid)
	//Cleanup
	defer con.RemoveClient()
	ac := appcontext.AppContext{}
	h, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
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
	kickoffRetryWatcher(g)
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
						// Wait for cluster to be reachable
						err = waitForClusterReady(c, cluster)
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
	return nil
}

// InstantiateComApp Instantiate Apps in Composite App
func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	go applyFnComApp(cid, appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiating},
		instantiateResource, addStatusTracker, true)
	return nil
}

// TerminateComApp Terminates Apps in Composite App
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	go applyFnComApp(cid, appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Terminating},
		terminateResource, deleteStatusTracker, false)
	return nil
}
