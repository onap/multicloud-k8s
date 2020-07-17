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
	"log"
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
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

func getRes(ac appcontext.AppContext, name string, app string, cluster string) ([]byte, error) {
	var byteRes []byte
	rh, err := ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		return nil, err
	}
	resval, err := ac.GetValue(rh)
	if err != nil {
		return nil, err
	}
	if resval != "" {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return nil, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		return nil, pkgerrors.Errorf("Resource value is nil %s", name)
	}
	return byteRes, nil
}

func terminateResource(ac appcontext.AppContext, c *kubeclient.Client, name string, app string, cluster string, label string) error {
	res, err := getRes(ac, name, app, cluster)
	if err != nil {
		return err
	}
	if err := c.Delete(res); err != nil {
		logutils.Error("Failed to delete res", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	logutils.Info("Deleted::", logutils.Fields{
		"cluster":  cluster,
		"resource": name,
	})
	return nil
}

func instantiateResource(ac appcontext.AppContext, c *kubeclient.Client, name string, app string, cluster string, label string) error {
	res, err := getRes(ac, name, app, cluster)
	if err != nil {
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
		logutils.Error("Failed to apply res", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	logutils.Info("Installed::", logutils.Fields{
		"cluster":  cluster,
		"resource": name,
	})
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

type fn func(ac appcontext.AppContext, client *kubeclient.Client, res string, app string, cluster string, label string) error

type statusfn func(client *kubeclient.Client, app string, cluster string, label string) error

func applyFnComApp(cid interface{}, con *connector.Connector, f fn, sfn statusfn, breakonError bool) error {
	ac := appcontext.AppContext{}
	g, _ := errgroup.WithContext(context.Background())
	_, err := ac.LoadAppContext(cid)
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

	for _, app := range appList["apporder"] {

		appName := app
		results := strings.Split(id.(string), "/")
		label := results[2] + "-" + app
		g.Go(func() error {
			clusterNames, err := ac.GetClusterNames(appName)
			if err != nil {
				return err
			}
			rg, _ := errgroup.WithContext(context.Background())
			for k := 0; k < len(clusterNames); k++ {
				cluster := clusterNames[k]
				err = status.StartClusterWatcher(cluster)
				if err != nil {
					log.Printf("Error starting Cluster Watcher %v: %v\n", cluster, err)
				}
				rg.Go(func() error {
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
					for i, res := range aov["resorder"] {
						err = f(ac, c, res, appName, cluster, label)
						if err != nil {
							logutils.Error("Error in resource %s: %v", logutils.Fields{
								"error":    err,
								"cluster":  cluster,
								"resource": res,
							})
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
					serr := sfn(c, appName, cluster, label)
					if serr != nil {
						logutils.Warn("Error handling status tracker", logutils.Fields{"error": serr})
					}
					return nil
				})
			}
			if err := rg.Wait(); err != nil {
				logutils.Error("Encountered error in App cluster", logutils.Fields{
					"error": err,
				})
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logutils.Error("Encountered error", logutils.Fields{
			"error": err,
		})
		return err
	}
	return nil
}

// InstantiateComApp Instantiate Apps in Composite App
func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	con := connector.Init(cid)
	err := applyFnComApp(cid, con, instantiateResource, addStatusTracker, true)
	if err != nil {
		logutils.Error("InstantiateComApp unsuccessful", logutils.Fields{"error": err})
		return err
	}
	//Cleanup
	con.RemoveClient()
	return nil
}

// TerminateComApp Terminates Apps in Composite App
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	con := connector.Init(cid)
	err := applyFnComApp(cid, con, terminateResource, deleteStatusTracker, false)
	if err != nil {
		logutils.Error("TerminateComApp unsuccessful", logutils.Fields{
			"error": err,
		})
		return err
	}
	//Cleanup
	con.RemoveClient()
	return nil
}
