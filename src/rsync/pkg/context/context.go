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
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/app"
	con "github.com/onap/multicloud-k8s/src/rsync/pkg/connector"
	res "github.com/onap/multicloud-k8s/src/rsync/pkg/resource"
	pkgerrors "github.com/pkg/errors"
)

type CompositeAppContext struct {
	cid interface{}
}

func deleteResource(clustername string, resname string, respath string) error {
	k8sClient := app.KubernetesClient{}
	err := k8sClient.Init(clustername, resname)
	if err != nil {
		log.Println("Init failed: " + err.Error())
		return err
	}

	var c con.KubernetesConnector
	c = &k8sClient
	var gp res.Resource
	err = gp.Delete(respath, resname, "default", c)
	if err != nil {
		log.Println("Delete resource failed: " + err.Error() + resname)
		return err
	}
	log.Println("Resource succesfully deleted", resname)
	return nil

}

func createResource(clustername string, resname string, resValue string) error {
	k8sClient := app.KubernetesClient{}
	err := k8sClient.Init(clustername, resname)
	if err != nil {
		log.Println("Client init failed: " + err.Error())
		return err
	}

	var c con.KubernetesConnector
	c = &k8sClient
	var gp res.Resource
	_, err = gp.Create(resValue, "default", c)
	if err != nil {
		log.Println("Create failed: " + err.Error() + resname)
		return err
	}
	log.Println("Resource succesfully created", resname)
	return nil

}

func terminateResource(ac appcontext.AppContext, res string, appname string, clustername string) error {

	rh, err := ac.GetResourceHandle(appname, clustername, res)
	if err != nil {
		return err
	}

	resval, err := ac.GetValue(rh)
	if err != nil {
		return err
	}

	if resval != "" {
		result := strings.Split(res, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil %s:", res)
		}
		err = deleteResource(clustername, result[0], resval.(string))
		if err != nil {
			return err
		}
	} else {
		return pkgerrors.Errorf("Resource value is nil")
	}

	return nil

}

func instantiateResource(ac appcontext.AppContext, name string, appname string, clustername string) error {
	rh, err := ac.GetResourceHandle(appname, clustername, name)
	if err != nil {
		return err
	}

	resval, err := ac.GetValue(rh)
	if err != nil {
		return err
	}

	if resval != "" {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		err = createResource(clustername, result[0], resval.(string))
		if err != nil {
			return err
		}
	} else {
		return pkgerrors.Errorf("Resource value is nil")
	}

	return nil

}

type fn func(ac appcontext.AppContext, res string, app string, cluster string) error

func applyFnComApp(cid interface{}, f fn, breakonError bool) error {
	var wg sync.WaitGroup
	ac := appcontext.AppContext{}

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
	for _, appName := range appList["apporder"] {
		wg.Add(1)
		go func(appName string) {
			defer wg.Done()
			clusterNames, err := ac.GetClusterNames(appName)
			if err != nil {
				return
			}
			for k := 0; k < len(clusterNames); k++ {
				wg.Add(1)
				go func(app, cluster string) {
					defer wg.Done()
					resorder, err := ac.GetResourceInstruction(app, cluster, "order")
					if err != nil {
						log.Printf("Resorder error %v", err)
						return
					}
					var aov map[string][]string
					json.Unmarshal([]byte(resorder.(string)), &aov)
					for _, res := range aov["resorder"] {
						logutils.Info("resorder list ----> ", logutils.Fields{
							"cluster":  cluster,
							"resource": res,
						})
						err = f(ac, res, app, cluster)
						if err != nil {
							if breakonError {
								log.Printf("Error in resource %s: %v", res, err)
								return
							}
						}
					}
					return
				}(appName, clusterNames[k])
			}
			return
		}(appName)
	}
	wg.Wait()
	return nil
}

func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	log.Printf("------ InstantiateComApp ----")
	err := applyFnComApp(cid, instantiateResource, false)
	if err != nil {
		log.Printf("InstantiateComApp unsuccessful")
		return err
	}
	return nil
}

// Delete all the resources for a given context
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	log.Printf("------ TerminateComApp ----")
	err := applyFnComApp(cid, terminateResource, false)
	if err != nil {
		log.Printf("TerminateComApp unsuccessful")
		return err
	}
	return nil
}
