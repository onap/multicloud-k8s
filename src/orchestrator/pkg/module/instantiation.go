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
	"fmt"
	gpic "github.com/onap/multicloud-k8s/src/orchestrator/pkg/gpic"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
)

// ManifestFileName is the name given to the manifest file in the profile package
const ManifestFileName = "manifest.yaml"

// GenericPlacementIntentName denotes the generic placement intent name
const GenericPlacementIntentName = "generic-placement-intent"

// SEPARATOR used while creating clusternames to store in etcd
const SEPARATOR = "+"

// InstantiationClient implements the InstantiationManager
type InstantiationClient struct {
	db InstantiationClientDbInfo
}

/*
InstantiationKey used in storing the contextid in the momgodb
It consists of
GenericPlacementIntentName,
ProjectName,
CompositeAppName,
CompositeAppVersion,
DeploymentIntentGroup
*/
type InstantiationKey struct {
	IntentName            string
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
	var found bool
	iList, err := NewIntentClient().GetAllIntents(p, ca, v, di)
	if err != nil {
		return gi, err
	}
	for _, eachMap := range iList.ListOfIntents {
		if gi, found := eachMap[GenericPlacementIntentName]; found {
			log.Info(":: Name of the generic-placement-intent ::", log.Fields{"GenPlmtIntent": gi})
			return gi, err
		}
	}
	if found == false {
		fmt.Println("generic-placement-intent not found !")
	}
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

	var appOrder []string

	// Add composite app using appContext
	for _, eachApp := range allApps {
		appOrder = append(appOrder, eachApp.Metadata.Name)
		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, overrideValues)

		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}

		log.Info(":: Resolved all the templates ::", log.Fields{"appName": eachApp.Metadata.Name, "SortedTemplate": sortedTemplates})

		resources, err := getResources(sortedTemplates)
		if err != nil {
			return pkgerrors.Wrapf(err, "Unable to get the resources for app :: %s", eachApp.Metadata.Name)
		}

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
	context.AddInstruction(compositeHandle, "app", "order", appOrder)
	//END: storing into etcd

	// BEGIN:: save the context in the orchestrator db record
	key := InstantiationKey{
		IntentName:            gIntent,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
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

	pl, err := getPrioritisedControllerList(p, ca, v, di)
	log.Info("Priority Based List ", log.Fields{"PlaCont::": pl.pPlaCont,
		"ActCont::": pl.pActCont})

	// END: Scheduler code

	log.Info(":: Done with instantiation... ::", log.Fields{"CompositeAppName": ca})
	return err
}
