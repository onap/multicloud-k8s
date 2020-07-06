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

	"golang.org/x/sync/errgroup"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/app"
	kubeclient "github.com/onap/multicloud-k8s/src/rsync/pkg/client"
	pkgerrors "github.com/pkg/errors"
)

type CompositeAppContext struct {
	cid interface{}
}

func createResource(clustername string, resname string, resValue []byte) error {
	k8sClient := app.KubernetesClient{}
	k, err := k8sClient.GetKubeConfigFile(clustername, resname)
	if err != nil {
		return err
	}
	c := kubeclient.New("", k)
	if err := c.Apply(resValue); err != nil {
		log.Printf("failed to apply the res %s", resname)
	}
	return nil
}

func deleteResource(clustername string, resname string, resValue []byte) error {
	k8sClient := app.KubernetesClient{}
	k, err := k8sClient.GetKubeConfigFile(clustername, resname)
	if err != nil {
		return err
	}
	c := kubeclient.New("", k)
	if err := c.Delete(resValue); err != nil {
		log.Printf("failed to apply the res %s", resname)
	}
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
		byteRes := []byte(fmt.Sprintf("%v", resval.(interface{})))
		err = deleteResource(clustername, result[0], byteRes)
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
		byteRes := []byte(fmt.Sprintf("%v", resval.(interface{})))
		err = createResource(clustername, result[0], byteRes)
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
	for _, app := range appList["apporder"] {

		appName := app
		g.Go(func() error {
			clusterNames, err := ac.GetClusterNames(appName)
			if err != nil {
				return err
			}
			rg, _ := errgroup.WithContext(context.Background())
			for k := 0; k < len(clusterNames); k++ {
				cluster := clusterNames[k]
				rg.Go(func() error {
					resorder, err := ac.GetResourceInstruction(appName, cluster, "order")
					if err != nil {
						log.Printf("Resorder error %v", err)
						return err
					}
					var aov map[string][]string
					json.Unmarshal([]byte(resorder.(string)), &aov)
					for _, res := range aov["resorder"] {
						err = f(ac, res, appName, cluster)
						if err != nil {
							if breakonError {
								log.Printf("Error in resource %s: %v", res, err)
								return err
							}
						}
						logutils.Info("Installed::", logutils.Fields{
							"cluster":  cluster,
							"resource": res,
						})
					}
					return nil
				})
			}
			if err := rg.Wait(); err != nil {
				fmt.Println("Encountered error in App cluster:", err)
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("Encountered error:", err)
		return err
	}
	return nil
}

// InstantiateComApp - Instantiate Apps in Composite App
func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	log.Printf("---- InstantiateComApp ----")
	err := applyFnComApp(cid, instantiateResource, true)
	if err != nil {
		log.Printf("InstantiateComApp unsuccessful")
		return err
	}
	return nil
}

// TerminateComApp Terminates Apps in Composite App
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	log.Printf("---- TerminateComApp ----")
	err := applyFnComApp(cid, terminateResource, false)
	if err != nil {
		log.Printf("TerminateComApp unsuccessful")
		return err
	}
	return nil
}
