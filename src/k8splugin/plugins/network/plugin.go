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
	"k8splugin/plugins/network/v1"
	"regexp"

	utils "k8splugin/internal"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

func extractData(data string) (vnfID, cniType, networkName string) {
	re := regexp.MustCompile("_")
	split := re.Split(data, -1)
	if len(split) != 3 {
		return
	}
	vnfID = split[0]
	cniType = split[1]
	networkName = split[2]
	return
}

// Create an ONAP Network object
func Create(data *utils.ResourceData, client kubernetes.Interface) (string, error) {
	network := &v1.OnapNetwork{}
	if _, err := utils.DecodeYAML(data.YamlFilePath, network); err != nil {
		return "", pkgerrors.Wrap(err, "Decode network object error")
	}

	config, err := network.DecodeConfig()
	if err != nil {
		return "", pkgerrors.Wrap(err, "Fail to decode network's configuration")
	}

	cniType := config["cnitype"].(string)
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

	return data.VnfId + "_" + cniType + "_" + name, nil
}

// List of Networks
func List(namespace string, kubeclient kubernetes.Interface) ([]string, error) {
	return nil, nil
}

// Delete an existing Network
func Delete(name string, namespace string, kubeclient kubernetes.Interface) error {
	_, cniType, networkName := extractData(name)
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

// Get an existing Network
func Get(name string, namespace string, kubeclient kubernetes.Interface) (string, error) {
	return "", nil
}
