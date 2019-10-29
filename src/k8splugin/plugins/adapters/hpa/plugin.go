/*
Copyright 2020 Intel Corporation.
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
	"errors"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"
)

// Compile check - does hpaPlacementPlugin implements the correct interface?
var _ plugin.PlacementAdapter = hpaPlacementPlugin{}

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable hpaPlacementPlugin

type hpaPlacementPlugin struct {
}

// Intent

// Validate and Store Intent
func (p hpaPlacementPlugin) StoreIntent(intent interface{}) error {
	return errors.New("Error")
}

// Validate and Modify Intent
func (p hpaPlacementPlugin) ModifyIntent(intentName string, intent interface{}) error {
	return errors.New("Error")
}

// Delete Intent
func (p hpaPlacementPlugin) DeleteIntent(intentName string) error {
	return errors.New("Error")
}

// Get Intent
func (p hpaPlacementPlugin) GetIntent(intentName string) (map[string]string, error) {
	//
	return make(map[string]string), errors.New("Error")
}

// Site

// Get Valid Sites to deploy workload given intents and sites
func (p hpaPlacementPlugin) GetSites(intentName string, profile rb.Profile, sites []string) (map[string][]string, error) {
	return make(map[string][]string), errors.New("Error")
}

// Node Features
// Store features for a cluster in mongoDB
func (p hpaPlacementPlugin) StoreFeatures(features interface{}) error {
	return errors.New("Error")
}

// Get Map of features on each node of cluster from mongoDB - map[<node-name string>][]<featureName string>
func (p hpaPlacementPlugin) GetFeaturesPerNode(clusterName string) (map[string][]string, error) {
	return make(map[string][]string), errors.New("Error")
}

// Get Array of features for a cluster from mongoDB - []<featureName string>
func (p hpaPlacementPlugin) GetFeatures(clusterName string) ([]string, error) {
	return make([]string, 0), errors.New("Error")
}

// Delete features for a cluster in mongoDB
func (p hpaPlacementPlugin) DeleteFeatures(clusterName string) error {
	return errors.New("Error")
}

// Modify features for a cluster in mongoDB
func (p hpaPlacementPlugin) ModifyFeatures(clusterName string, Features interface{}) error {
	return errors.New("Error")
}
