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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	gpic "github.com/onap/multicloud-k8s/src/orchestrator/pkg/gpic"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
	"io/ioutil"
	"log"
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
			log.Printf("::Name of the generic-placement-intent:: %s", gi)
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

	log.Println("Processing App.. ", appName)

	var sortedTemplates []helm.KubernetesResourceTemplate

	aC, err := NewAppClient().GetAppContent(appName, p, ca, v)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprint("Not finding the content of app:: ", appName))
	}
	appContent, err := base64.StdEncoding.DecodeString(aC.FileContent)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}
	log.Println("Got the app content..")

	appPC, err := NewAppProfileClient().GetAppProfileContentByApp(p, ca, v, cp, appName)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprintf("Not finding the appProfileContent for:: %s", appName))
	}
	appProfileContent, err := base64.StdEncoding.DecodeString(appPC.Profile)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Println("Got the app Profile content ...")

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

	log.Printf("The len of the sortedTemplates :: %d", len(sortedTemplates))

	return sortedTemplates, err
}

// resource consists of name of reource
type resource struct {
	name        string
	filecontent []byte
}

// getResources shall take in the sorted templates and output the resources
// which consists of name(name+kind) and filecontent
func getResources(st []helm.KubernetesResourceTemplate) ([]resource, error) {
	var resources []resource
	for _, t := range st {
		yamlStruct, err := utils.ExtractYamlParameters(t.FilePath)
		yamlFile, err := ioutil.ReadFile(t.FilePath)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Failed to get the resources..")
		}
		n := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind

		resources = append(resources, resource{name: n, filecontent: yamlFile})
		log.Printf("Added resource :: % s :: into resource-order", n)
	}
	return resources, nil
}

func addResourcesToCluster(ct appcontext.AppContext, ch interface{}, resources []resource, resourceOrder []string) error {
	for _, resource := range resources {

		resourceOrder = append(resourceOrder, resource.name)
		_, err := ct.AddResource(ch, resource.name, resource.filecontent)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf("Error :: %s", cleanuperr.Error())
				log.Printf("Error Cleaning up AppContext after add resource failure. resource-name:: %s", resource.name)
			}
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource.name)
		}
		_, err = ct.AddInstruction(ch, "resource", "order", resourceOrder)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf("Error :: %s", cleanuperr.Error())
				log.Printf("Error Cleaning up AppContext after add instruction failure. resource-name:: %s", resource.name)
			}
			return pkgerrors.Wrapf(err, "Error adding instruction for resource ::%s to AppContext", resource.name)
		}
	}
	return nil
}

func addClustersToAppContext(l gpic.Clusters, ct appcontext.AppContext, appHandle interface{}, resources []resource) error {
	for _, c := range l.ClustersWithName {
		p := c.ProviderName
		n := c.ClusterName
		var resourceOrder []string
		clusterhandle, err := ct.AddCluster(appHandle, p+SEPARATOR+n)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf("Error :: %s", cleanuperr.Error())
				log.Printf("Error Cleaning up AppContext after add cluster failure. cluster-provider:: %s, cluster-name:: %s", p, n)
			}
			return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
		}

		err = addResourcesToCluster(ct, clusterhandle, resources, resourceOrder)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s and name::%s) to AppContext", p, n)
		}
	}
	return nil
}

/*
verifyResources method is just to check if the resource handles are correctly saved.
*/

func verifyResources(l gpic.Clusters, ct appcontext.AppContext, resources []resource, appName string)  error {
	for _, c := range l.ClustersWithName {
		p := c.ProviderName
		n := c.ClusterName
		cn := p+SEPARATOR+n
		for _, res := range resources {

			rh, err := ct.GetResourceHandle(appName, cn, res.name)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting resoure handle for resource :: %s, app:: %s, cluster :: %s", appName, res.name, cn)
			}
			log.Printf("AppName :: %s, Cluster :: %s, Resource :: %s", appName, cn, res.name)
			log.Printf("Resource handle:: %v", rh)
		}

	}

	return nil
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
	log.Printf("The name of the GenPlacIntent:: %s", gIntent)

	log.Printf("dIGrp :: %s, releaseName :: %s and cp :: %s \n", dIGrp.MetaData.Name, rName, cp)
	allApps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the apps")
	}

	// Make an app context for the compositeApp
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}

	var appOrder []string

	// Add composite app using appContext
	for _, eachApp := range allApps {
		appOrder = append(appOrder, eachApp.Metadata.Name)
		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, overrideValues)

		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}
		log.Printf("Resolved all the templates for app :: %s under the compositeApp...", eachApp.Metadata.Name)
		log.Printf("sortedTemplates :: %v ", sortedTemplates)

		resources, err := getResources(sortedTemplates)
		if err != nil {
			return pkgerrors.Wrapf(err, "Unable to get the resources for app :: %s", eachApp.Metadata.Name)
		}

		specData, err := NewAppIntentClient().GetAllIntentsByApp(eachApp.Metadata.Name, p, ca, v, gIntent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents for app")
		}
		listOfClusters, err := gpic.IntentResolver(specData.Intent)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the intents resolved for app")
		}
		log.Printf("::listOfClusters:: %v", listOfClusters)

		//BEGIN: storing into etcd
		// Add an app to the app context
		apphandle, err := context.AddApp(compositeHandle, eachApp.Metadata.Name)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf("Error :: %s", cleanuperr.Error())
				log.Printf("Error Cleaning up AppContext compositeApp failure. AppName:: %s", eachApp.Metadata.Name)
			}
			return pkgerrors.Wrap(err, "Error adding App to AppContext")
		}
		err = addClustersToAppContext(listOfClusters, context, apphandle, resources)
		if err!=nil {
			log.Printf("Error while adding cluster and resources to app:: %s", eachApp.Metadata.Name)
		}
		err = verifyResources(listOfClusters, context, resources, eachApp.Metadata.Name)
		if err != nil {
			log.Printf("Error while verifying resources in app:: %s", eachApp.Metadata.Name)
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
			log.Printf("Error cleaning up AppContext while saving context in the db for GPIntent:: %s, Project:: %s, CompositeApp:: %s, CompositeAppVerion::%s, DeploymentIntentGroup:: %s", gIntent, p, ca, v, di)
		}
		return pkgerrors.Wrap(err, "Error adding AppContext to DB")
	}
	// END:: save the context in the orchestrator db record

	log.Printf("Done with instantiation...")
	return err
}
