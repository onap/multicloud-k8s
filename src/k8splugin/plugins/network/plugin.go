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
	"io/ioutil"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"

	"k8splugin/krd"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
)

type NetworkFile struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name    string `json:"name"`
		Cnitype string `json:"cnitype"`
	} `json:"metadata"`
}

// nwLoadedPlugins stores references to the stored plugins for networking
var nwLoadedPlugins = map[string]*plugin.Plugin{}

// ReadNetworkFile reads the network yaml
var ReadNetworkFile = func(path string) (NetworkFile, error) {
	var networkFile NetworkFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return networkFile, pkgerrors.Wrap(err, "Network YAML file does not exist")
	}

	log.Println("Reading Network YAML: " + path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return networkFile, pkgerrors.Wrap(err, "Network YAML file read error")
	}
	err = yaml.Unmarshal(yamlFile, &networkFile)
	if err != nil {
		return networkFile, pkgerrors.Wrap(err, "Network YAML file unmarshal error")
	}
	log.Printf("network:\n%v", networkFile)

	return networkFile, nil
}

// LoadPlugins loads all the compiled .so plugins and call initialize function
func nwLoadPlugins(initData *krd.InitData) error {
	pluginsDir := os.Getenv("NW_PLUGINS_DIR")
	err := filepath.Walk(pluginsDir,
		func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, ".so") {
				p, err := plugin.Open(path)
				if err != nil {
					return pkgerrors.Cause(err)
				}
				nwLoadedPlugins[info.Name()[:len(info.Name())-3]] = p
				symCreateResourceFunc, err := p.Lookup("Initialize")
				if err != nil {
					return pkgerrors.Wrap(err, "Error fetching Initialize function")
				}
				internalResourceName, err := symCreateResourceFunc.(func(*krd.InitData) (string, error))(initData)
				if err != nil {
					return pkgerrors.Wrap(err, "Error in plugin "+internalResourceName)
				}
			}
			return err
		})
	if err != nil {
		return err
	}

	return nil
}

// Initialize Plugin
func Initialize(data *krd.InitData) (string, error) {

	log.Printf("Intiialize Network Plugin")

	// Load Network Plugins
	err := nwLoadPlugins(data)
	if err != nil {
		return "Network", pkgerrors.Cause(err)
	}

	return "Network", nil
}

func extract_data(data string) (vnf_id, cniType, networkName string) {

	re := regexp.MustCompile("_")
	split := re.Split(data, -1)
	if len(split) != 3 {
		return
	}
	vnf_id = split[0]
	cniType = split[1]
	networkName = split[2]
	return
}

// Create a Network
func Create(data *krd.ResourceData, client kubernetes.Interface) (string, error) {

	log.Printf("Create virtual Network")

	namespace := data.Namespace
	if namespace == "" {
		namespace = "default"
	}

	networkFile, err := ReadNetworkFile(data.YamlFilePath)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error while reading File: "+data.YamlFilePath)
	}

	cniType := networkFile.Metadata.Cnitype

	// Call Submodule
	typePlugin, ok := nwLoadedPlugins[cniType]
	if !ok {
		return "", pkgerrors.New("No plugin for resource " + cniType + " found")
	}

	symCreateResourceFunc, err := typePlugin.Lookup("CreateNetwork")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error fetching "+cniType+" plugin")
	}

	internalResourceName, err := symCreateResourceFunc.(func(*krd.ResourceData, kubernetes.Interface) (string, error))(
		data, client)

	if err != nil {
		return "", pkgerrors.Wrap(err, "Error in plugin "+internalResourceName+" plugin")
	}

	networkName := data.VnfId + "_" + cniType + "_" + internalResourceName
	return networkName, nil
}

// List of Networks
func List(namespace string, kubeclient kubernetes.Interface) ([]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	return nil, nil
}

// Delete an existing Network
func Delete(name string, namespace string, kubeclient kubernetes.Interface) error {
	if namespace == "" {
		namespace = "default"
	}

	// Extract CNI Type
	_, cniType, networkName := extract_data(name)

	// Call Submodule
	typePlugin, ok := nwLoadedPlugins[cniType]
	if !ok {
		return pkgerrors.New("No plugin for resource " + cniType + " found")
	}
	symCreateResourceFunc, err := typePlugin.Lookup("DeleteNetwork")
	if err != nil {
		return pkgerrors.Wrap(err, "Error fetching "+cniType+" DeleteNetwork")
	}
	err = symCreateResourceFunc.(func(string, kubernetes.Interface) error)(
		networkName, kubeclient)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in plugin "+cniType+" DeleteNetwork ")
	}
	return nil
}

// Get an existing Network
func Get(name string, namespace string, kubeclient kubernetes.Interface) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	// Extract CNI Type
	_, cniType, networkName := extract_data(name)

	// Call Submodule
	typePlugin, ok := nwLoadedPlugins[cniType]
	if !ok {
		return "", pkgerrors.New("No plugin for resource " + cniType + " found")
	}
	symCreateResourceFunc, err := typePlugin.Lookup("GetNetwork")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error fetching "+cniType+" GetNetwork")
	}
	internalResource, err := symCreateResourceFunc.(func(string, kubernetes.Interface) (string, error))(
		networkName, kubeclient)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error in plugin "+internalResource+" GetNetwork")
	}
	return internalResource, nil
}
