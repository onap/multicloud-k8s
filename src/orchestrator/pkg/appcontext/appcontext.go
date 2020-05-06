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

package appcontext

import (
	"fmt"
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/rtcontext"
	pkgerrors "github.com/pkg/errors"
	//"log"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
)

type AppContext struct {
	initDone bool
	rtcObj   rtcontext.RunTimeContext
	rtc      rtcontext.Rtcontext
}

// Init app context
func (ac *AppContext) InitAppContext() (interface{}, error) {
	ac.rtcObj = rtcontext.RunTimeContext{}
	ac.rtc = &ac.rtcObj
	return ac.rtc.RtcInit()
}

// Load app context that was previously created
func (ac *AppContext) LoadAppContext(cid interface{}) (interface{}, error) {
	ac.rtcObj = rtcontext.RunTimeContext{}
	ac.rtc = &ac.rtcObj
	return ac.rtc.RtcLoad(cid)
}

// Create a new context and returns the handle
func (ac *AppContext) CreateCompositeApp() (interface{}, error) {
	h, err := ac.rtc.RtcCreate()
	if err != nil {
		return nil, err
	}
	return h, nil
}

// Deletes the entire context
func (ac *AppContext) DeleteCompositeApp() error {
	h, err := ac.rtc.RtcGet()
	if err != nil {
		return err
	}
	err = ac.rtc.RtcDeletePrefix(h)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handles for a given composite app context
func (ac *AppContext) GetCompositeApp() (interface{}, error) {
	h, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}
	return h, nil
}

//Add app to the context under composite app
func (ac *AppContext) AddApp(handle interface{}, appname string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(handle, "app", appname)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added app handle ::", log.Fields{"AppHandle":h})
	return h, nil
}

//Delete app from the context and everything underneth
func (ac *AppContext) DeleteApp(handle interface{}) error {
	err := ac.rtc.RtcDeletePrefix(handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handle for a given app
func (ac *AppContext) GetAppHandle(appname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}

	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	apph := fmt.Sprintf("%v", rh) + "app/" + appname + "/"
	hs, err := ac.rtc.RtcGetHandles(apph)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == apph {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given app")
}

//Add cluster to the context under app
func (ac *AppContext) AddCluster(handle interface{}, clustername string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(handle, "cluster", clustername)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added cluster handle ::", log.Fields{"ClusterHandler":h})
	return h, nil
}

//Delete cluster from the context and everything underneth
func (ac *AppContext) DeleteCluster(handle interface{}) error {
	err := ac.rtc.RtcDeletePrefix(handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the handle for a given app and cluster
func (ac *AppContext) GetClusterHandle(appname string, clustername string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	ach := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/"
	hs, err := ac.rtc.RtcGetHandles(ach)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == ach {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given cluster")
}

//Returns a list of all clusters for a given app
func (ac *AppContext) GetClusterNames(appname string) ([]string, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}

	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/"
	hs, err := ac.rtc.RtcGetHandles(prefix)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting handles for %v", prefix)
	}
	var cs []string
	for _, h := range hs {
		hstr := fmt.Sprintf("%v", h)
		ks := strings.Split(hstr, prefix)
		for _, k := range ks {
			ck := strings.Split(k, "/")
			if len(ck) == 2 && ck[1] == "" {
				cs = append(cs, ck[0])
			}
		}
	}
	return cs, nil
}

//Add resource under app and cluster
func (ac *AppContext) AddResource(handle interface{}, resname string, value []byte) (interface{}, error) {
	h, err := ac.rtc.RtcAddResource(handle, resname, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added resource handle ::", log.Fields{"ResourceHandler":h})

	return h, nil
}

//Delete resource given the handle
func (ac *AppContext) DeleteResource(handle interface{}) error {
	err := ac.rtc.RtcDeletePair(handle)
	if err != nil {
		return err
	}
	return nil
}

//Return the hanlde for given app, cluster and resource name
func (ac *AppContext) GetResourceHandle(appname string, clustername string, resname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/" + resname + "/"
	hs, err := ac.rtc.RtcGetHandles(acrh)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == acrh {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for the given resource")
}

//Update the resource value usign the given handle
func (ac *AppContext) UpdateResourceValue(handle interface{}, value []byte) error {
	return ac.rtc.RtcUpdateValue(handle, value)
}

//Add instruction under given handle and type
func (ac *AppContext) AddInstruction(handle interface{}, level string, insttype string, value interface{}) (interface{}, error) {
	if !(insttype == "order" || insttype == "dependency") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	if !(level == "app" || level == "resource") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction level")
	}
	h, err := ac.rtc.RtcAddInstruction(handle, level, insttype, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added instruction handle ::", log.Fields{"InstructionHandler":h})
	return h, nil
}

//Delete instruction under gievn handle
func (ac *AppContext) DeleteInstruction(handle interface{}) error {
	err := ac.rtc.RtcDeletePair(handle)
	if err != nil {
		return err
	}
	return nil
}

//Returns the app instruction for a given instruction type
func (ac *AppContext) GetAppInstruction(insttype string) (interface{}, error) {
	if !(insttype == "order" || insttype == "dependency") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf("%v", rh) + "app/" + "instruction/" + insttype + "/"
	var v string
	err = ac.rtc.RtcGetValue(s, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

//Update the instruction usign the given handle
func (ac *AppContext) UpdateInstructionValue(handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(handle, value)
}

//Returns the resource instruction for a given instruction type
func (ac *AppContext) GetResourceInstruction(appname string, clustername string, insttype string) (interface{}, error) {
	if !(insttype == "order" || insttype == "dependency") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/instruction/" + insttype + "/"
	var v string
	err = ac.rtc.RtcGetValue(s, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

//Return all the handles under the composite app
func (ac *AppContext) GetAllHandles(handle interface{}) ([]interface{}, error) {
	hs, err := ac.rtc.RtcGetHandles(handle)
	if err != nil {
		return nil, err
	}
	return hs, nil
}

//Returns the value for a given handle
func (ac *AppContext) GetValue(handle interface{}) (interface{}, error) {
	var v string
	err := ac.rtc.RtcGetValue(handle, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
