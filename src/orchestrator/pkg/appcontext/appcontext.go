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

	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/rtcontext"
	pkgerrors "github.com/pkg/errors"
)

// metaPrefix used for denoting clusterMeta level
const metaGrpPREFIX = "!@#metaGrp"

type AppContext struct {
	initDone bool
	rtcObj   rtcontext.RunTimeContext
	rtc      rtcontext.Rtcontext
}

// AppContextStatus represents the current status of the appcontext
//	Instantiating - instantiate has been invoked and is still in progress
//	Instantiated - instantiate has completed
//	Terminating - terminate has been invoked and is still in progress
//	Terminated - terminate has completed
//	InstantiateFailed - the instantiate action has failed
//	TerminateFailed - the terminate action has failed
type AppContextStatus struct {
	Status StatusValue
}
type StatusValue string
type statuses struct {
	Instantiating     StatusValue
	Instantiated      StatusValue
	Terminating       StatusValue
	Terminated        StatusValue
	InstantiateFailed StatusValue
	TerminateFailed   StatusValue
}

var AppContextStatusEnum = &statuses{
	Instantiating:     "Instantiating",
	Instantiated:      "Instantiated",
	Terminating:       "Terminating",
	Terminated:        "Terminated",
	InstantiateFailed: "InstantiateFailed",
	TerminateFailed:   "TerminateFailed",
}

// CompositeAppMeta consists of projectName, CompositeAppName,
// CompositeAppVersion, ReleaseName. This shall be used for
// instantiation of a compositeApp
type CompositeAppMeta struct {
	Project               string `json:"Project"`
	CompositeApp          string `json:"CompositeApp"`
	Version               string `json:"Version"`
	Release               string `json:"Release"`
	DeploymentIntentGroup string `json:"DeploymentIntentGroup"`
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

// CreateCompositeApp method returns composite app handle as interface.
func (ac *AppContext) CreateCompositeApp() (interface{}, error) {
	h, err := ac.rtc.RtcCreate()
	if err != nil {
		return nil, err
	}
	log.Info(":: CreateCompositeApp ::", log.Fields{"CompositeAppHandle": h})
	return h, nil
}

// AddCompositeAppMeta adds the meta data associated with a composite app
func (ac *AppContext) AddCompositeAppMeta(meta interface{}) error {
	err := ac.rtc.RtcAddMeta(meta)
	if err != nil {
		return err
	}
	return nil
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
func (ac *AppContext) GetCompositeAppHandle() (interface{}, error) {
	h, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}
	return h, nil
}

// GetLevelHandle returns the handle for the supplied level at the given handle.
// For example, to get the handle of the 'status' level at a given handle.
func (ac *AppContext) GetLevelHandle(handle interface{}, level string) (interface{}, error) {
	ach := fmt.Sprintf("%v%v/", handle, level)
	hs, err := ac.rtc.RtcGetHandles(ach)
	if err != nil {
		return nil, err
	}
	for _, v := range hs {
		if v == ach {
			return v, nil
		}
	}
	return nil, pkgerrors.Errorf("No handle was found for level %v", level)
}

//Add app to the context under composite app
func (ac *AppContext) AddApp(handle interface{}, appname string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(handle, "app", appname)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added app handle ::", log.Fields{"AppHandle": h})
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

// AddCluster helps to add cluster to the context under app. It takes in the app handle and clusterName as value.
func (ac *AppContext) AddCluster(handle interface{}, clustername string) (interface{}, error) {
	h, err := ac.rtc.RtcAddLevel(handle, "cluster", clustername)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added cluster handle ::", log.Fields{"ClusterHandler": h})
	return h, nil
}

// AddClusterMetaGrp adds the meta info of groupNumber to which a cluster belongs.
// It takes in cluster handle and groupNumber as arguments
func (ac *AppContext) AddClusterMetaGrp(ch interface{}, gn string) error {
	mh, err := ac.rtc.RtcAddOneLevel(ch, metaGrpPREFIX, gn)
	if err != nil {
		return err
	}
	log.Info(":: Added cluster meta handle ::", log.Fields{"ClusterMetaHandler": mh})
	return nil
}

// DeleteClusterMetaGrpHandle deletes the group number to which the cluster belongs, it takes in the cluster handle.
func (ac *AppContext) DeleteClusterMetaGrpHandle(ch interface{}) error {
	err := ac.rtc.RtcDeletePrefix(ch)
	if err != nil {
		return err
	}
	log.Info(":: Deleted cluster meta handle ::", log.Fields{"ClusterMetaHandler": ch})
	return nil
}

/*
GetClusterMetaHandle takes in appName and ClusterName as string arguments and return the ClusterMetaHandle as string
*/
func (ac *AppContext) GetClusterMetaHandle(app string, cluster string) (string, error) {
	if app == "" {
		return "", pkgerrors.Errorf("Not a valid run time context app name")
	}
	if cluster == "" {
		return "", pkgerrors.Errorf("Not a valid run time context cluster name")
	}

	ch, err := ac.GetClusterHandle(app, cluster)
	if err != nil {
		return "", err
	}
	cmh := fmt.Sprintf("%v", ch) + metaGrpPREFIX + "/"
	return cmh, nil

}

/*
GetClusterGroupMap shall take in appName and return a map showing the grouping among the clusters.
sample output of "GroupMap" :{"1":["cluster_provider1+clusterName3","cluster_provider1+clusterName5"],"2":["cluster_provider2+clusterName4","cluster_provider2+clusterName6"]}
*/
func (ac *AppContext) GetClusterGroupMap(an string) (map[string][]string, error) {
	cl, err := ac.GetClusterNames(an)
	if err != nil {
		log.Info(":: Unable to fetch clusterList for app ::", log.Fields{"AppName ": an})
		return nil, err
	}
	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	var gmap = make(map[string][]string)
	for _, cn := range cl {
		s := fmt.Sprintf("%v", rh) + "app/" + an + "/cluster/" + cn + "/" + metaGrpPREFIX + "/"
		var v string
		err = ac.rtc.RtcGetValue(s, &v)
		if err != nil {
			log.Info(":: No group number for cluster  ::", log.Fields{"cluster": cn, "Reason": err})
			continue
		}
		gn := fmt.Sprintf("%v", v)
		log.Info(":: GroupNumber retrieved  ::", log.Fields{"GroupNumber": gn})

		cl, found := gmap[gn]
		if found == false {
			cl = make([]string, 0)
		}
		cl = append(cl, cn)
		gmap[gn] = cl
	}
	return gmap, nil
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
func (ac *AppContext) AddResource(handle interface{}, resname string, value interface{}) (interface{}, error) {
	h, err := ac.rtc.RtcAddResource(handle, resname, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added resource handle ::", log.Fields{"ResourceHandler": h})

	return h, nil
}

//Return the handle for given app, cluster and resource name
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

//Update the resource value using the given handle
func (ac *AppContext) UpdateResourceValue(handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(handle, value)
}

//Return the handle for given app, cluster and resource name
func (ac *AppContext) GetResourceStatusHandle(appname string, clustername string, resname string) (interface{}, error) {
	if appname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context app name")
	}
	if clustername == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context cluster name")
	}
	if resname == "" {
		return nil, pkgerrors.Errorf("Not a valid run time context resource name")
	}

	rh, err := ac.rtc.RtcGet()
	if err != nil {
		return nil, err
	}

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/resource/" + resname + "/status/"
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

//Add instruction under given handle and type
func (ac *AppContext) AddInstruction(handle interface{}, level string, insttype string, value interface{}) (interface{}, error) {
	if !(insttype == "order" || insttype == "dependency") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction type")
	}
	if !(level == "app" || level == "resource" || level == "subresource") {
		return nil, pkgerrors.Errorf("Not a valid app context instruction level")
	}
	h, err := ac.rtc.RtcAddInstruction(handle, level, insttype, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added instruction handle ::", log.Fields{"InstructionHandler": h})
	return h, nil
}

//Delete instruction under given handle
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

// AddLevelValue for holding a state object at a given level
// will make a handle with an appended "<level>/" to the key
func (ac *AppContext) AddLevelValue(handle interface{}, level string, value interface{}) (interface{}, error) {
	h, err := ac.rtc.RtcAddOneLevel(handle, level, value)
	if err != nil {
		return nil, err
	}
	log.Info(":: Added handle ::", log.Fields{"Handle": h})

	return h, nil
}

// GetClusterStatusHandle returns the handle for cluster status for a given app and cluster
func (ac *AppContext) GetClusterStatusHandle(appname string, clustername string) (interface{}, error) {
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

	acrh := fmt.Sprintf("%v", rh) + "app/" + appname + "/cluster/" + clustername + "/status/"
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

//UpdateStatusValue updates the status value with the given handle
func (ac *AppContext) UpdateStatusValue(handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(handle, value)
}

//UpdateValue updates the state value with the given handle
func (ac *AppContext) UpdateValue(handle interface{}, value interface{}) error {
	return ac.rtc.RtcUpdateValue(handle, value)
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
	var v interface{}
	err := ac.rtc.RtcGetValue(handle, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetCompositeAppMeta returns the meta data associated with the compositeApp
// Its return type is CompositeAppMeta
func (ac *AppContext) GetCompositeAppMeta() (CompositeAppMeta, error) {
	mi, err := ac.rtcObj.RtcGetMeta()

	if err != nil {
		return CompositeAppMeta{}, pkgerrors.Errorf("Failed to get compositeApp meta")
	}
	datamap, ok := mi.(map[string]interface{})
	if ok == false {
		return CompositeAppMeta{}, pkgerrors.Errorf("Failed to cast meta interface to compositeApp meta")
	}

	p := fmt.Sprintf("%v", datamap["Project"])
	ca := fmt.Sprintf("%v", datamap["CompositeApp"])
	v := fmt.Sprintf("%v", datamap["Version"])
	rn := fmt.Sprintf("%v", datamap["Release"])
	dig := fmt.Sprintf("%v", datamap["DeploymentIntentGroup"])

	return CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rn, DeploymentIntentGroup: dig}, nil
}
