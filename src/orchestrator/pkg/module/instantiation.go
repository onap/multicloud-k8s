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

package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	rb "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	gpic "github.com/onap/multicloud-k8s/src/orchestrator/pkg/gpic"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
)

// ManifestFileName is the name given to the manifest file in the profile package
const ManifestFileName = "manifest.yaml"

// GenericPlacementIntentName denotes the generic placement intent name
const GenericPlacementIntentName = "genericPlacementIntent"

// SEPARATOR used while creating clusternames to store in etcd
const SEPARATOR = "+"

// InstantiationClient implements the InstantiationManager
type InstantiationClient struct {
	db InstantiationClientDbInfo
}

type ClusterAppStatus struct {
	Cluster string
	App     string
	Status  rb.ResourceBundleStatus
}

type StatusData struct {
	Data []ClusterAppStatus
}

/*
InstantiationKey used in storing the contextid in the momgodb
It consists of
ProjectName,
CompositeAppName,
CompositeAppVersion,
DeploymentIntentGroup
*/
type InstantiationKey struct {
	Project               string
	CompositeApp          string
	Version               string
	DeploymentIntentGroup string
}

// InstantiationManager is an interface which exposes the
// InstantiationManager functionalities
type InstantiationManager interface {
	//ApproveInstantiation(p string, ca string, v string, di string) (error)
	Instantiate(p string, ca string, v string, di string) error
	Status(p string, ca string, v string, di string) (StatusData, error)
	Terminate(p string, ca string, v string, di string) error
}

// InstantiationClientDbInfo consists of storeName and tagContext
type InstantiationClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for Instantiationclient documents
	tagContext string // attribute key name for context object in App Context
}

// NewInstantiationClient returns an instance of InstantiationClient
func NewInstantiationClient() *InstantiationClient {
	return &InstantiationClient{
		db: InstantiationClientDbInfo{
			storeName:  "orchestrator",
			tagContext: "contextid",
		},
	}
}

// TODO
//ApproveInstantiation approves an instantiation
// func (c InstantiationClient) ApproveInstantiation(p string, ca string, v string, di string) (error){
// }

func getOverrideValuesByAppName(ov []OverrideValues, a string) map[string]string {
	for _, eachOverrideVal := range ov {
		if eachOverrideVal.AppName == a {
			return eachOverrideVal.ValuesObj
		}
	}
	return map[string]string{}
}

/*
findGenericPlacementIntent takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentIntentName
and returns the name of the genericPlacementIntentName. Returns empty value if string not found.
*/
func findGenericPlacementIntent(p, ca, v, di string) (string, error) {
	var gi string
	iList, err := NewIntentClient().GetAllIntents(p, ca, v, di)
	if err != nil {
		return gi, err
	}
	for _, eachMap := range iList.ListOfIntents {
		if gi, found := eachMap[GenericPlacementIntentName]; found {
			log.Info(":: Name of the generic-placement-intent found ::", log.Fields{"GenPlmtIntent": gi})
			return gi, nil
		}
	}
	log.Info(":: generic-placement-intent not found ! ::", log.Fields{"Searched for GenPlmtIntent": GenericPlacementIntentName})
	return gi, pkgerrors.New("Generic-placement-intent not found")
}

// GetSortedTemplateForApp returns the sorted templates.
//It takes in arguments - appName, project, compositeAppName, releaseName, compositeProfileName, array of override values
func GetSortedTemplateForApp(appName, p, ca, v, rName, cp string, overrideValues []OverrideValues) ([]helm.KubernetesResourceTemplate, error) {

	log.Info(":: Processing App ::", log.Fields{"appName": appName})

	var sortedTemplates []helm.KubernetesResourceTemplate

	aC, err := NewAppClient().GetAppContent(appName, p, ca, v)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprint("Not finding the content of app:: ", appName))
	}
	appContent, err := base64.StdEncoding.DecodeString(aC.FileContent)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app content.. ::", log.Fields{"appName": appName})

	appPC, err := NewAppProfileClient().GetAppProfileContentByApp(p, ca, v, cp, appName)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprintf("Not finding the appProfileContent for:: %s", appName))
	}
	appProfileContent, err := base64.StdEncoding.DecodeString(appPC.Profile)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app Profile content .. ::", log.Fields{"appName": appName})

	overrideValuesOfApp := getOverrideValuesByAppName(overrideValues, appName)
	//Convert override values from map to array of strings of the following format
	//foo=bar
	overrideValuesOfAppStr := []string{}
	if overrideValuesOfApp != nil {
		for k, v := range overrideValuesOfApp {
			overrideValuesOfAppStr = append(overrideValuesOfAppStr, k+"="+v)
		}
	}

	sortedTemplates, err = helm.NewTemplateClient("", "default", rName,
		ManifestFileName).Resolve(appContent,
		appProfileContent, overrideValuesOfAppStr,
		appName)

	log.Info(":: Total no. of sorted templates ::", log.Fields{"len(sortedTemplates):": len(sortedTemplates)})

	return sortedTemplates, err
}

/*
Instantiate methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible for template resolution, intent
resolution, creation and saving of context for saving into etcd.
*/
func (c InstantiationClient) Instantiate(p string, ca string, v string, di string) error {

	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}
	rName := dIGrp.Spec.Version //rName is releaseName
	overrideValues := dIGrp.Spec.OverrideValuesObj
	cp := dIGrp.Spec.Profile

	gIntent, err := findGenericPlacementIntent(p, ca, v, di)
	if err != nil {
		return err
	}

	log.Info(":: The name of the GenPlacIntent ::", log.Fields{"GenPlmtIntent": gIntent})
	log.Info(":: DeploymentIntentGroup, ReleaseName, CompositeProfile ::", log.Fields{"dIGrp": dIGrp.MetaData.Name, "releaseName": rName, "cp": cp})

	allApps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the apps")
	}

	cca, err := makeAppContextForCompositeApp(p, ca, v, rName)
	if err != nil {
		return err
	}
	context := cca.context
	ctxval := cca.ctxval
	compositeHandle := cca.compositeAppHandle

	var appOrderInstr struct {
		Apporder []string `json:"apporder"`
	}

	var appDepInstr struct {
		Appdep map[string]string `json:"appdependency"`
	}
	appdep := make(map[string]string)

	// Add composite app using appContext
	for _, eachApp := range allApps {
		appOrderInstr.Apporder = append(appOrderInstr.Apporder, eachApp.Metadata.Name)
		appdep[eachApp.Metadata.Name] = "go"

		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, overrideValues)

		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}

		log.Info(":: Resolved all the templates ::", log.Fields{"appName": eachApp.Metadata.Name, "SortedTemplate": sortedTemplates})

		resources, err := getResources(sortedTemplates)
		if err != nil {
			return pkgerrors.Wrapf(err, "Unable to get the resources for app :: %s", eachApp.Metadata.Name)
		}

		statusResource, err := getStatusResource(ctxval.(string), eachApp.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrapf(err, "Unable to generate the status resource for app :: %s", eachApp.Metadata.Name)
		}
		resources = append(resources, statusResource)

		specData, err := NewAppIntentClient().GetAllIntentsByApp(eachApp.Metadata.Name, p, ca, v, gIntent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents for app")
		}
		// listOfClusters shall have both mandatoryClusters and optionalClusters where the app needs to be installed.
		listOfClusters, err := gpic.IntentResolver(specData.Intent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents resolved for app")
		}

		log.Info(":: listOfClusters ::", log.Fields{"listOfClusters": listOfClusters})

		//BEGIN: storing into etcd
		// Add an app to the app context
		apphandle, err := context.AddApp(compositeHandle, eachApp.Metadata.Name)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext compositeApp failure ::", log.Fields{"Error": cleanuperr.Error(), "AppName": eachApp.Metadata.Name})
			}
			return pkgerrors.Wrap(err, "Error adding App to AppContext")
		}
		err = addClustersToAppContext(listOfClusters, context, apphandle, resources)
		if err != nil {
			log.Info(":: Error while adding cluster and resources to app ::", log.Fields{"Error": err.Error(), "AppName": eachApp.Metadata.Name})
		}
		err = verifyResources(listOfClusters, context, resources, eachApp.Metadata.Name)
		if err != nil {
			log.Info(":: Error while verifying resources in app ::", log.Fields{"Error": err.Error(), "AppName": eachApp.Metadata.Name})
		}

	}
	jappOrderInstr, _ := json.Marshal(appOrderInstr)
	appDepInstr.Appdep = appdep
	jappDepInstr, _ := json.Marshal(appDepInstr)
	context.AddInstruction(compositeHandle, "app", "order", string(jappOrderInstr))
	context.AddInstruction(compositeHandle, "app", "dependency", string(jappDepInstr))
	//END: storing into etcd

	// BEGIN:: save the context in the orchestrator db record
	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagContext, ctxval)
	if err != nil {
		cleanuperr := context.DeleteCompositeApp()
		if cleanuperr != nil {

			log.Info(":: Error Cleaning up AppContext while saving context in the db for GPIntent ::", log.Fields{"Error": cleanuperr.Error(), "GPIntent": gIntent, "DeploymentIntentGroup": di, "CompositeApp": ca, "CompositeAppVersion": v, "Project": p})
		}
		return pkgerrors.Wrap(err, "Error adding AppContext to DB")
	}
	// END:: save the context in the orchestrator db record

	// BEGIN: scheduler code

	pl, mapOfControllers, err := getPrioritizedControllerList(p, ca, v, di)
	if err != nil {
		return err
	}
	log.Info("Priority Based List ", log.Fields{"PlacementControllers::": pl.pPlaCont,
		"ActionControllers::": pl.pActCont, "mapOfControllers::": mapOfControllers})

	err = callGrpcForControllerList(pl.pPlaCont, mapOfControllers, ctxval)
	if err != nil {
		return err
	}

	err = deleteExtraClusters(allApps, context)
	if err != nil {
		return err
	}

	err = callGrpcForControllerList(pl.pActCont, mapOfControllers, ctxval)
	if err != nil {
		return err
	}

	// END: Scheduler code

	// BEGIN : Rsync code
	err = callRsync(ctxval)
	if err != nil {
		return err
	}
	// END : Rsyc code

	log.Info(":: Done with instantiation... ::", log.Fields{"CompositeAppName": ca})
	return err
}

/*
Status takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible obtaining the status of
the deployment, which is made available in the appcontext.
*/
func (c InstantiationClient) Status(p string, ca string, v string, di string) (StatusData, error) {

	ac, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupContext(di, p, ca, v)
	if err != nil {
		return StatusData{}, pkgerrors.Wrap(err, "deploymentIntentGroup not found "+di)
	}

	// Get all apps in this composite app
	allApps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return StatusData{}, pkgerrors.Wrap(err, "Not finding the apps")
	}

	var diStatus StatusData
	diStatus.Data = make([]ClusterAppStatus, 0)

	// Loop through each app and get the status data for each cluster in the app
	for _, app := range allApps {
		// Get the clusters in the appcontext for this app
		clusters, err := ac.GetClusterNames(app.Metadata.Name)
		if err != nil {
			log.Info(":: No clusters for app ::", log.Fields{"AppName": app.Metadata.Name})
			continue
		}

		for _, cluster := range clusters {
			handle, err := ac.GetStatusHandle(app.Metadata.Name, cluster)
			if err != nil {
				log.Info(":: No status handle for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app.Metadata.Name, "Error": err})
				continue
			}
			statusValue, err := ac.GetValue(handle)
			if err != nil {
				log.Info(":: No status value for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app.Metadata.Name, "Error": err})
				continue
			}
			log.Info(":: STATUS VALUE ::", log.Fields{"statusValue": statusValue})
			var statusData ClusterAppStatus
			err = json.Unmarshal([]byte(statusValue.(string)), &statusData.Status)
			if err != nil {
				log.Info(":: Error unmarshaling status value for cluster, app ::",
					log.Fields{"Cluster": cluster, "AppName": app.Metadata.Name, "Error": err})
				continue
			}
			statusData.Cluster = cluster
			statusData.App = app.Metadata.Name
			log.Info(":: STATUS DATA ::", log.Fields{"status": statusData})

			diStatus.Data = append(diStatus.Data, statusData)
		}
	}

	return diStatus, nil
}

/*
Terminate takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and calls rsync to terminate.
*/
func (c InstantiationClient) Terminate(p string, ca string, v string, di string) error {

	//ac, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupContext(di, p, ca, v)
	_, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupContext(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "deploymentIntentGroup not found "+di)
	}

	// TODO - make call to rsync to terminate the composite app deployment
	//        will leave the appcontext in place for clean up later
	//        so monitoring status can be performed

	return nil
}
