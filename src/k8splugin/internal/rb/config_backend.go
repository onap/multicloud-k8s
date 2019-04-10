/*
 * Copyright 2018 Intel Corporation, Inc
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

package rb

import (
	"bytes"
	"encoding/json"
	"k8splugin/internal/db"
	"k8splugin/internal/helm"
	"log"
	"strconv"
	"strings"

	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
)

//ConfigStore contains the values that will be stored in the database
type configVersionDBContent struct {
	ConfigNew  Config `json:"config-new"`
	ConfigPrev Config `json:"config-prev"`
	Action     string `json:"action"` // CRUD opration for this config
}

//ConfigStore to Store the Config
type ConfigStore struct {
	rbName      string
	rbVersion   string
	profileName string
	configName  string
}

//ConfigVersionStore to Store the Versions of the Config
type ConfigVersionStore struct {
	rbName      string
	rbVersion   string
	profileName string
}

type configResourceList struct {
	retmap  map[string][]string
	profile Profile
	action  string
}

type profileDataManager struct {
	profileLockMap  map[string]*sync.Mutex
	resourceChannel map[string](chan configResourceList)
	lock            sync.Mutex
}

const (
	storeName  = "config"
	tagCounter = "counter"
	tagVersion = "configversion"
	tagConfig  = "configdata"
)

var profileData = profileDataManager{
	lock:            sync.Mutex{},
	profileLockMap:  map[string]*sync.Mutex{},
	resourceChannel: map[string]chan configResourceList{},
}

// Construct key for storing data
func constructKey(strs ...string) string {

	var sb strings.Builder
	sb.WriteString("onapk8s")
	sb.WriteString("/")
	sb.WriteString(storeName)
	sb.WriteString("/")
	for _, str := range strs {
		sb.WriteString(str)
		sb.WriteString("/")
	}
	return sb.String()

}

// Create an entry for the config in the database
func (c ConfigStore) createConfig(p Config) error {

	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagConfig, p.ConfigName)
	_, err := db.Etcd.Get(cfgKey)
	if err == nil {
		return pkgerrors.Wrap(err, "Config DB Entry Already exists")
	}
	configValue, err := db.Serialize(p)
	if err != nil {
		return pkgerrors.Wrap(err, "Serialize Config Value")
	}
	err = db.Etcd.Put(cfgKey, configValue)
	if err != nil {
		return pkgerrors.Wrap(err, "Config DB Entry")
	}
	return nil
}

// Update the config entry in the database. Updates with the new value
// Returns the previous value of the Config
func (c ConfigStore) updateConfig(p Config) (Config, error) {

	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagConfig, p.ConfigName)
	value, err := db.Etcd.Get(cfgKey)
	configPrev := Config{}
	if err == nil {
		// If updating Config after rollback then previous config may not exist
		err = db.DeSerialize(string(value), &configPrev)
		if err != nil {
			return Config{}, pkgerrors.Wrap(err, "DeSerialize Config Value")
		}
	}
	configValue, err := db.Serialize(p)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Serialize Config Value")
	}
	err = db.Etcd.Put(cfgKey, configValue)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Config DB Entry")
	}
	return configPrev, nil
}

// Read the config entry in the database
func (c ConfigStore) getConfig() (Config, error) {
	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagConfig, c.configName)
	value, err := db.Etcd.Get(cfgKey)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config DB Entry")
	}
	//value is a byte array
	if value != nil {
		cfg := Config{}
		err = db.DeSerialize(string(value), &cfg)
		if err != nil {
			return Config{}, pkgerrors.Wrap(err, "Unmarshaling Config Value")
		}
		return cfg, nil
	}
	return Config{}, pkgerrors.Wrap(err, "Get Config DB Entry")
}

// Delete the config entry in the database
func (c ConfigStore) deleteConfig() (Config, error) {

	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagConfig, c.configName)
	value, err := db.Etcd.Get(cfgKey)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Config DB Entry Not found")
	}
	configPrev := Config{}
	err = db.DeSerialize(string(value), &configPrev)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "DeSerialize Config Value")
	}

	err = db.Etcd.Delete(cfgKey)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Config DB Entry")
	}
	return configPrev, nil
}

// Create a version for the configuration. If previous config provided that is also stored
func (c ConfigVersionStore) createConfigVersion(configNew, configPrev Config, action string) (uint, error) {

	version, err := c.incrementVersion()

	if err != nil {
		return 0, pkgerrors.Wrap(err, "Get Next Version")
	}
	versionKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagVersion, strconv.Itoa(int(version)))

	var cs configVersionDBContent
	cs.Action = action
	cs.ConfigNew = configNew
	cs.ConfigPrev = configPrev

	configValue, err := db.Serialize(cs)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Serialize Config Value")
	}
	err = db.Etcd.Put(versionKey, configValue)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Create Config DB Entry")
	}
	return version, nil
}

// Delete current version of the configuration. Configuration always deleted from top
func (c ConfigVersionStore) deleteConfigVersion() error {

	counter, err := c.getCurrentVersion()

	if err != nil {
		return pkgerrors.Wrap(err, "Get Next Version")
	}
	versionKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagVersion, strconv.Itoa(int(counter)))

	err = db.Etcd.Delete(versionKey)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Config DB Entry")
	}
	err = c.decrementVersion()
	if err != nil {
		return pkgerrors.Wrap(err, "Decrement Version")
	}
	return nil
}

// Read the specified version of the configuration and return its prev and current value.
// Also returns the action for the config version
func (c ConfigVersionStore) getConfigVersion(version uint) (Config, Config, string, error) {

	versionKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagVersion, strconv.Itoa(int(version)))
	configBytes, err := db.Etcd.Get(versionKey)
	if err != nil {
		return Config{}, Config{}, "", pkgerrors.Wrap(err, "Get Config Version ")
	}

	if configBytes != nil {
		pr := configVersionDBContent{}
		err = db.DeSerialize(string(configBytes), &pr)
		if err != nil {
			return Config{}, Config{}, "", pkgerrors.Wrap(err, "DeSerialize Config Version")
		}
		return pr.ConfigNew, pr.ConfigPrev, pr.Action, nil
	}
	return Config{}, Config{}, "", pkgerrors.Wrap(err, "Invalid data ")
}

// Get the counter for the version
func (c ConfigVersionStore) getCurrentVersion() (uint, error) {

	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagCounter)

	value, err := db.Etcd.Get(cfgKey)
	if err != nil {
		// Counter not started yet, 0 is invalid value
		return 0, nil
	}
	index, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Invalid counter")
	}
	return uint(index), nil
}

// Update the counter for the version
func (c ConfigVersionStore) updateVersion(counter uint) error {

	cfgKey := constructKey(c.rbName, c.rbVersion, c.profileName, tagCounter)
	err := db.Etcd.Put(cfgKey, strconv.Itoa(int(counter)))
	if err != nil {
		return pkgerrors.Wrap(err, "Counter DB Entry")
	}
	return nil
}

// Increment the version counter
func (c ConfigVersionStore) incrementVersion() (uint, error) {

	counter, err := c.getCurrentVersion()
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Get Next Counter Value")
	}
	//This is done while Profile lock is taken
	counter++
	err = c.updateVersion(counter)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Store Next Counter Value")
	}

	return counter, nil
}

// Decrement the version counter
func (c ConfigVersionStore) decrementVersion() error {

	counter, err := c.getCurrentVersion()
	if err != nil {
		return pkgerrors.Wrap(err, "Get Next Counter Value")
	}
	//This is done while Profile lock is taken
	counter--
	err = c.updateVersion(counter)
	if err != nil {
		return pkgerrors.Wrap(err, "Store Next Counter Value")
	}

	return nil
}

// Apply Config
func applyConfig(rbName, rbVersion, profileName string, p Config, pChannel chan configResourceList, action string) error {

	// Get Template and Resolve the template with values
	crl, err := resolve(rbName, rbVersion, profileName, p)
	if err != nil {
		return pkgerrors.Wrap(err, "Resolve Config")
	}
	crl.action = action
	// Send the configResourceList to the channel. Using select for non-blocking channel
	select {
	case pChannel <- crl:
		log.Printf("Message Sent to goroutine %v", crl.profile)
	default:
	}

	return nil
}

// Per Profile Go routine to apply the configuration to Cloud Region
func scheduleResources(c chan configResourceList) {
	// Keep thread running
	for {
		data := <-c
		//TODO: ADD Check to see if Application running
		switch {
		case data.action == "POST":
			log.Printf("[scheduleResources]: POST %v %v", data.profile, data.retmap)
			//TODO: Needs to add code to call Kubectl create
		case data.action == "PUT":
			log.Printf("[scheduleResources]: PUT %v %v", data.profile, data.retmap)
			//TODO: Needs to add code to call Kubectl apply
		case data.action == "DELETE":
			log.Printf("[scheduleResources]: DELETE %v %v", data.profile, data.retmap)
			//TODO: Needs to add code to call Kubectl delete

		}
	}
}

//Resolve returns the path where the helm chart merged with
//configuration overrides resides.
var resolve = func(rbName, rbVersion, profileName string, p Config) (configResourceList, error) {

	var retMap map[string][]string

	def, err := NewConfigTemplateClient().Download(rbName, rbVersion, p.TemplateName)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Downloading Template")
	}

	profile, err := NewProfileClient().Get(rbName, rbVersion, profileName)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Reading  Profile Data")
	}

	chartBasePath, err := ExtractTarBall(bytes.NewBuffer(def))
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Extracting Template")
	}

	var chartName string
	finfo, err := ioutil.ReadDir(chartBasePath)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Detecting chart name")
	}

	//Store the first directory with Chart.yaml found as the chart name
	for _, f := range finfo {
		if f.IsDir() {
			//Check if Chart.yaml exists
			if _, err = os.Stat(filepath.Join(chartBasePath, f.Name(), "Chart.yaml")); err == nil {
				chartName = f.Name()
				break
			}
		}
	}

	if chartName == "" {
		return configResourceList{}, pkgerrors.New("Invalid template no Chart.yaml file found")
	}

	b, err := json.Marshal(p.Values)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Error Marshalling config data")
	}
	data, err := yaml.JSONToYAML(b)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "JSON to YAML")
	}

	//Create a temp file in the system temp folder
	outputfile, err := ioutil.TempFile("", "helm-config-values-")
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Got error creating temp file")
	}
	_, err = outputfile.Write([]byte(data))
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Got error writting temp file")
	}
	defer outputfile.Close()

	helmClient := helm.NewTemplateClient(profile.KubernetesVersion,
		profile.Namespace,
		profile.ReleaseName)

	chartPath := filepath.Join(chartBasePath, chartName)
	retMap, err = helmClient.GenerateKubernetesArtifacts(chartPath,
		[]string{outputfile.Name()},
		nil)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Generate final k8s yaml")
	}
	crl := configResourceList{
		retmap:  retMap,
		profile: profile,
	}

	return crl, nil
}

// Get the Mutex for the Profile
func getProfileData(key string) (*sync.Mutex, chan configResourceList) {
	profileData.lock.Lock()
	defer profileData.lock.Unlock()
	_, ok := profileData.profileLockMap[key]
	if !ok {
		profileData.profileLockMap[key] = &sync.Mutex{}
	}
	_, ok = profileData.resourceChannel[key]
	if !ok {
		profileData.resourceChannel[key] = make(chan configResourceList)
		go scheduleResources(profileData.resourceChannel[key])
	}
	return profileData.profileLockMap[key], profileData.resourceChannel[key]
}
