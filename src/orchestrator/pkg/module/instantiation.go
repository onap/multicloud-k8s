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
	"fmt"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils/helm"

	"github.com/onap/multicloud-k8s/src/orchestrator/utils/types"
	pkgerrors "github.com/pkg/errors"

	"encoding/base64"
	"log"
)

// ManifestFileName is the name given to the manifest file in the profile package
const ManifestFileName = "manifest.yaml"

// InstantiationClient implements the InstantiationManager
type InstantiationClient struct {
	storeName   string
	tagMetaData string
}

// InstantiationManager is an interface which exposes the
// InstantiationManager functionalities
type InstantiationManager interface {
	//ApproveInstantiation(p string, ca string, v string, di string) (error)
	Instantiate(p string, ca string, v string, di string) error
}

// NewInstantiationClient returns an instance of InstantiationClient
func NewInstantiationClient() *InstantiationClient {
	return &InstantiationClient{
		storeName:   "orchestrator",
		tagMetaData: "instantiation",
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

// GetSortedTemplateForApp returns the sorted templates.
//It takes in arguments - appName, project, compositeAppName, releaseName, compositeProfileName, array of override values
func GetSortedTemplateForApp(appName, p, ca, v, rName, cp string, overrideValues []OverrideValues) ([]types.KubernetesResourceTemplate, error) {

	log.Println("Processing App.. ", appName)

	var sortedTemplates []types.KubernetesResourceTemplate

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
		rName, appName)

	log.Printf("The len of the sortedTemplates :: %d", len(sortedTemplates))

	return sortedTemplates, err
}

// Instantiate methods takes in project
func (c InstantiationClient) Instantiate(p string, ca string, v string, di string) error {

	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}
	rName := dIGrp.Spec.Version //rName is releaseName
	overrideValues := dIGrp.Spec.OverrideValuesObj
	cp := dIGrp.Spec.Profile

	log.Printf("dIGrp :: %s, releaseName :: %s and cp :: %s \n", dIGrp.MetaData.Name, rName, cp)
	allApps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the apps")
	}
	for _, eachApp := range allApps {
		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, overrideValues)
		if err != nil {
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}
		log.Printf("Resolved all the templates for app :: %s under the compositeApp...", eachApp.Metadata.Name)
		log.Printf("sortedTemplates :: %v ", sortedTemplates)
	}
	log.Printf("Done with instantiation...")
	return err
}
