/*
Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	v1 "k8splugin/plugins/network/v1"
	"regexp"

	utils "k8splugin/internal"
	"k8splugin/internal/app"
	"k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable networkPlugin

type networkPlugin struct {
}

func extractData(data string) (cniType, networkName string) {
	re := regexp.MustCompile("_")
	split := re.Split(data, -1)
	if len(split) != 3 {
		return
	}
	cniType = split[1]
	networkName = split[2]
	return
}

// Create an ONAP Network object
func (p networkPlugin) Create(yamlFilePath string, namespace string, client *app.KubernetesClient) (string, error) {
	network := &v1.OnapNetwork{}
	if _, err := utils.DecodeYAML(yamlFilePath, network); err != nil {
		return "", pkgerrors.Wrap(err, "Decode network object error")
	}

	cniType := network.Spec.CniType
	typePlugin, ok := utils.LoadedPlugins[cniType+"-network"]
	if !ok {
		return "", pkgerrors.New("No plugin for resource " + cniType + " found")
	}

	symCreateNetworkFunc, err := typePlugin.Lookup("CreateNetwork")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error fetching "+cniType+" plugin")
	}

	name, err := symCreateNetworkFunc.(func(*v1.OnapNetwork) (string, error))(network)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error during the creation for "+cniType+" plugin")
	}

	return cniType + "_" + name, nil
}

// Get a Network
func (p networkPlugin) Get(resource helm.KubernetesResource, namespace string, client *app.KubernetesClient) (string, error) {
	return "", nil
}

// List of Networks
func (p networkPlugin) List(gvk schema.GroupVersionKind, namespace string,
	client *app.KubernetesClient) ([]helm.KubernetesResource, error) {

	return nil, nil
}

// Delete an existing Network
func (p networkPlugin) Delete(resource helm.KubernetesResource, namespace string, client *app.KubernetesClient) error {
	cniType, networkName := extractData(resource.Name)
	typePlugin, ok := utils.LoadedPlugins[cniType+"-network"]
	if !ok {
		return pkgerrors.New("No plugin for resource " + cniType + " found")
	}

	symDeleteNetworkFunc, err := typePlugin.Lookup("DeleteNetwork")
	if err != nil {
		return pkgerrors.Wrap(err, "Error fetching "+cniType+" plugin")
	}

	if err := symDeleteNetworkFunc.(func(string) error)(networkName); err != nil {
		return pkgerrors.Wrap(err, "Error during the deletion for "+cniType+" plugin")
	}

	return nil
}
