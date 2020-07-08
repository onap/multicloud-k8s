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
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/app"
	con "github.com/onap/multicloud-k8s/src/rsync/pkg/connector"
	res "github.com/onap/multicloud-k8s/src/rsync/pkg/resource"
	status "github.com/onap/multicloud-k8s/src/rsync/pkg/status"
	pkgerrors "github.com/pkg/errors"
)

type CompositeAppContext struct {
	cid            interface{}
	appsorder      string
	appsdependency string
	appsmap        []instMap
}
type clusterInfo struct {
	name          string
	resorder      string
	resdependency string
	ressmap       []instMap
}
type instMap struct {
	name     string
	depinfo  string
	status   string
	rerr     error
	clusters []clusterInfo
}

func getInstMap(order string, dependency string, level string) ([]instMap, error) {

	if order == "" {
		return nil, pkgerrors.Errorf("Not a valid order value")
	}
	if dependency == "" {
		return nil, pkgerrors.Errorf("Not a valid dependency value")
	}

	if !(level == "app" || level == "res") {
		return nil, pkgerrors.Errorf("Not a valid level name given to create map")
	}

	var aov map[string]interface{}
	json.Unmarshal([]byte(order), &aov)

	s := fmt.Sprintf("%vorder", level)
	appso := aov[s].([]interface{})
	var instmap = make([]instMap, len(appso))

	var adv map[string]interface{}
	json.Unmarshal([]byte(dependency), &adv)
	s = fmt.Sprintf("%vdependency", level)
	appsd := adv[s].(map[string]interface{})
	for i, u := range appso {
		instmap[i] = instMap{u.(string), appsd[u.(string)].(string), "none", nil, nil}
	}

	return instmap, nil
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

func createResource(clustername string, resname string, respath string, label string) error {
	k8sClient := app.KubernetesClient{}
	err := k8sClient.Init(clustername, resname)
	if err != nil {
		log.Println("Client init failed: " + err.Error())
		return err
	}

	var c con.KubernetesConnector
	c = &k8sClient
	var gp res.Resource
	_, err = gp.Create(respath, "default", label, c)
	if err != nil {
		log.Println("Create failed: " + err.Error() + resname)
		return err
	}
	log.Println("Resource succesfully created", resname)
	return nil

}

func terminateResource(ac appcontext.AppContext, resmap instMap, appname string, clustername string) error {

	rh, err := ac.GetResourceHandle(appname, clustername, resmap.name)
	if err != nil {
		return err
	}

	resval, err := ac.GetValue(rh)
	if err != nil {
		return err
	}

	if resval != "" {
		result := strings.Split(resmap.name, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil")
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

func instantiateResource(ac appcontext.AppContext, resmap instMap, appname string, clustername string, label string) error {
	rh, err := ac.GetResourceHandle(appname, clustername, resmap.name)
	if err != nil {
		return err
	}

	resval, err := ac.GetValue(rh)
	if err != nil {
		return err
	}

	if resval != "" {
		result := strings.Split(resmap.name, "+")
		if result[0] == "" {
			return pkgerrors.Errorf("Resource name is nil")
		}
		err = createResource(clustername, result[0], resval.(string), label)
		if err != nil {
			return err
		}
	} else {
		return pkgerrors.Errorf("Resource value is nil")
	}

	return nil

}

func terminateResources(ac appcontext.AppContext, ressmap []instMap, appname string, clustername string) error {
	var wg sync.WaitGroup
	var chans = make([]chan int, len(ressmap))
	for l := range chans {
		chans[l] = make(chan int)
	}
	for i := 0; i < len(ressmap); i++ {
		wg.Add(1)
		go func(index int) {
			if ressmap[index].depinfo == "go" {
				ressmap[index].status = "start"
			} else {
				ressmap[index].status = "waiting"
				c := <-chans[index]
				if c != index {
					ressmap[index].status = "error"
					ressmap[index].rerr = pkgerrors.Errorf("channel does not match")
					wg.Done()
					return
				}
				ressmap[index].status = "start"
			}
			ressmap[index].rerr = terminateResource(ac, ressmap[index], appname, clustername)
			ressmap[index].status = "done"
			waitstr := fmt.Sprintf("wait on %v", ressmap[index].name)
			for j := 0; j < len(ressmap); j++ {
				if ressmap[j].depinfo == waitstr {
					chans[j] <- j
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	for k := 0; k < len(ressmap); k++ {
		if ressmap[k].rerr != nil {
			return pkgerrors.Errorf("error during resources termination")
		}
	}
	return nil

}

func instantiateResources(ac appcontext.AppContext, ressmap []instMap, appname string, clustername string) error {
	var wg sync.WaitGroup
	var chans = make([]chan int, len(ressmap))
	cid, _ := ac.GetCompositeAppHandle()

	results := strings.Split(cid.(string), "/")
	label := results[2] + "-" + appname

	for l := range chans {
		chans[l] = make(chan int)
	}
	for i := 0; i < len(ressmap); i++ {
		wg.Add(1)
		go func(index int) {
			if ressmap[index].depinfo == "go" {
				ressmap[index].status = "start"
			} else {
				ressmap[index].status = "waiting"
				c := <-chans[index]
				if c != index {
					ressmap[index].status = "error"
					ressmap[index].rerr = pkgerrors.Errorf("channel does not match")
					wg.Done()
					return
				}
				ressmap[index].status = "start"
			}
			ressmap[index].rerr = instantiateResource(ac, ressmap[index], appname, clustername, label)
			ressmap[index].status = "done"
			waitstr := fmt.Sprintf("wait on %v", ressmap[index].name)
			for j := 0; j < len(ressmap); j++ {
				if ressmap[j].depinfo == waitstr {
					chans[j] <- j
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	for k := 0; k < len(ressmap); k++ {
		if ressmap[k].rerr != nil {
			return pkgerrors.Errorf("error during resources instantiation")
		}
	}
	return nil

}

func terminateApp(ac appcontext.AppContext, appmap instMap) error {

	for i := 0; i < len(appmap.clusters); i++ {
		err := terminateResources(ac, appmap.clusters[i].ressmap, appmap.name,
			appmap.clusters[i].name)
		if err != nil {
			return err
		}
	}
	log.Println("Termination of app done: " + appmap.name)

	return nil

}

func instantiateApp(ac appcontext.AppContext, appmap instMap) error {

	for i := 0; i < len(appmap.clusters); i++ {
		err := instantiateResources(ac, appmap.clusters[i].ressmap, appmap.name,
			appmap.clusters[i].name)
		if err != nil {
			return err
		}
		err = status.StartClusterWatcher(appmap.clusters[i].name)
		if err != nil {
			log.Printf("Error starting Cluster Watcher %v: %v\n", appmap.clusters[i], err)
		}
	}
	log.Println("Instantiation of app done: " + appmap.name)
	return nil

}

func instantiateApps(ac appcontext.AppContext, appsmap []instMap) error {
	var wg sync.WaitGroup
	var chans = make([]chan int, len(appsmap))
	for l := range chans {
		chans[l] = make(chan int)
	}
	for i := 0; i < len(appsmap); i++ {
		wg.Add(1)
		go func(index int) {
			if appsmap[index].depinfo == "go" {
				appsmap[index].status = "start"
			} else {
				appsmap[index].status = "waiting"
				c := <-chans[index]
				if c != index {
					appsmap[index].status = "error"
					appsmap[index].rerr = pkgerrors.Errorf("channel does not match")
					wg.Done()
					return
				}
				appsmap[index].status = "start"
			}
			appsmap[index].rerr = instantiateApp(ac, appsmap[index])
			appsmap[index].status = "done"
			waitstr := fmt.Sprintf("wait on %v", appsmap[index].name)
			for j := 0; j < len(appsmap); j++ {
				if appsmap[j].depinfo == waitstr {
					chans[j] <- j
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	for k := 0; k < len(appsmap); k++ {
		if appsmap[k].rerr != nil {
			return pkgerrors.Errorf("error during apps instantiation")
		}
	}
	return nil

}

func (instca *CompositeAppContext) InstantiateComApp(cid interface{}) error {
	ac := appcontext.AppContext{}

	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
	}
	instca.cid = cid

	appsorder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	instca.appsorder = appsorder.(string)
	appsdependency, err := ac.GetAppInstruction("dependency")
	if err != nil {
		return err
	}
	instca.appsdependency = appsdependency.(string)
	instca.appsmap, err = getInstMap(instca.appsorder, instca.appsdependency, "app")
	if err != nil {
		return err
	}

	for j := 0; j < len(instca.appsmap); j++ {
		clusternames, err := ac.GetClusterNames(instca.appsmap[j].name)
		if err != nil {
			return err
		}
		instca.appsmap[j].clusters = make([]clusterInfo, len(clusternames))
		for k := 0; k < len(clusternames); k++ {
			instca.appsmap[j].clusters[k].name = clusternames[k]
			resorder, err := ac.GetResourceInstruction(
				instca.appsmap[j].name, clusternames[k], "order")
			if err != nil {
				return err
			}
			instca.appsmap[j].clusters[k].resorder = resorder.(string)

			resdependency, err := ac.GetResourceInstruction(
				instca.appsmap[j].name, clusternames[k], "dependency")
			if err != nil {
				return err
			}
			instca.appsmap[j].clusters[k].resdependency = resdependency.(string)

			instca.appsmap[j].clusters[k].ressmap, err = getInstMap(
				instca.appsmap[j].clusters[k].resorder,
				instca.appsmap[j].clusters[k].resdependency, "res")
			if err != nil {
				return err
			}
		}
	}
	err = instantiateApps(ac, instca.appsmap)
	if err != nil {
		return err
	}

	return nil
}

// Delete all the apps
func terminateApps(ac appcontext.AppContext, appsmap []instMap) error {
	var wg sync.WaitGroup
	var chans = make([]chan int, len(appsmap))
	for l := range chans {
		chans[l] = make(chan int)
	}
	for i := 0; i < len(appsmap); i++ {
		wg.Add(1)
		go func(index int) {
			if appsmap[index].depinfo == "go" {
				appsmap[index].status = "start"
			} else {
				appsmap[index].status = "waiting"
				c := <-chans[index]
				if c != index {
					appsmap[index].status = "error"
					appsmap[index].rerr = pkgerrors.Errorf("channel does not match")
					wg.Done()
					return
				}
				appsmap[index].status = "start"
			}
			appsmap[index].rerr = terminateApp(ac, appsmap[index])
			appsmap[index].status = "done"
			waitstr := fmt.Sprintf("wait on %v", appsmap[index].name)
			for j := 0; j < len(appsmap); j++ {
				if appsmap[j].depinfo == waitstr {
					chans[j] <- j
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	for k := 0; k < len(appsmap); k++ {
		if appsmap[k].rerr != nil {
			return pkgerrors.Errorf("error during apps instantiation")
		}
	}
	return nil

}

// Delete all the resources for a given context
func (instca *CompositeAppContext) TerminateComApp(cid interface{}) error {
	ac := appcontext.AppContext{}

	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
	}
	instca.cid = cid

	appsorder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	instca.appsorder = appsorder.(string)
	appsdependency, err := ac.GetAppInstruction("dependency")
	if err != nil {
		return err
	}
	instca.appsdependency = appsdependency.(string)
	instca.appsmap, err = getInstMap(instca.appsorder, instca.appsdependency, "app")
	if err != nil {
		return err
	}

	for j := 0; j < len(instca.appsmap); j++ {
		clusternames, err := ac.GetClusterNames(instca.appsmap[j].name)
		if err != nil {
			return err
		}
		instca.appsmap[j].clusters = make([]clusterInfo, len(clusternames))
		for k := 0; k < len(clusternames); k++ {
			instca.appsmap[j].clusters[k].name = clusternames[k]
			resorder, err := ac.GetResourceInstruction(
				instca.appsmap[j].name, clusternames[k], "order")
			if err != nil {
				return err
			}
			instca.appsmap[j].clusters[k].resorder = resorder.(string)

			resdependency, err := ac.GetResourceInstruction(
				instca.appsmap[j].name, clusternames[k], "dependency")
			if err != nil {
				return err
			}
			instca.appsmap[j].clusters[k].resdependency = resdependency.(string)

			instca.appsmap[j].clusters[k].ressmap, err = getInstMap(
				instca.appsmap[j].clusters[k].resorder,
				instca.appsmap[j].clusters[k].resdependency, "res")
			if err != nil {
				return err
			}
		}
	}
	err = terminateApps(ac, instca.appsmap)
	if err != nil {
		return err
	}

	return nil

}
