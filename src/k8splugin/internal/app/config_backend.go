/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
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

package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	"github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
)

//ConfigStore contains the values that will be stored in the database
type configVersionDBContent struct {
	ConfigNew  Config                     `json:"config-new"`
	ConfigPrev Config                     `json:"config-prev"`
	Action     string                     `json:"action"` // CRUD opration for this config
	Resources  []KubernetesConfigResource `json:"resources"`
}

//ConfigStore to Store the Config
type ConfigStore struct {
	instanceID string
	configName string
}

//ConfigVersionStore to Store the Versions of the Config
type ConfigVersionStore struct {
	instanceID string
	configName string
}

type KubernetesConfigResource struct {
	Resource helm.KubernetesResource `json:"resource"`
	Status   string                  `json:"status"`
}

type configResourceList struct {
	resourceTemplates []helm.KubernetesResourceTemplate
	resources         []KubernetesConfigResource
	updatedResources  chan []KubernetesConfigResource
	profile           rb.Profile
	action            string
}

type profileDataManager struct {
	profileLockMap  map[string]*sync.Mutex
	resourceChannel map[string](chan configResourceList)
	sync.Mutex
}

const (
	storeName  = "config"
	tagCounter = "counter"
	tagVersion = "configversion"
	tagName    = "configtag"
	tagConfig  = "configdata"
)

var profileData = profileDataManager{
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

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig, p.ConfigName)
	_, err = db.Etcd.Get(cfgKey)
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

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig, p.ConfigName)
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
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig, c.configName)
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

// Read the config entry in the database
func (c ConfigStore) getConfigList() ([]Config, error) {
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return []Config{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig)
	values, err := db.Etcd.GetValues(cfgKey)
	if err != nil {
		return []Config{}, pkgerrors.Wrap(err, "Get Config DB List")
	}
	//value is a byte array
	if values != nil {
		result := make([]Config, 0)
		for _, value := range values {
			cfg := Config{}
			err = db.DeSerialize(string(value), &cfg)
			if err != nil {
				return []Config{}, pkgerrors.Wrap(err, "Unmarshaling Config Value")
			}
			if cfg.ConfigName == "" {
				continue
			}
			result = append(result, cfg)
		}
		return result, nil
	}
	return []Config{}, pkgerrors.Wrap(err, "Get Config DB List")
}

// Delete the config entry in the database
func (c ConfigStore) deleteConfig() (Config, error) {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig, c.configName)
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

//Cleanup stored data in etcd before instance is being deleted
func (c ConfigVersionStore) cleanupIstanceTags(configName string) error {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}

	versionKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagVersion, configName)
	err = db.Etcd.DeletePrefix(versionKey)
	if err != nil {
		log.Printf("Deleting versions of instance failed: %s", err.Error())
	}

	counterKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagCounter, configName)
	err = db.Etcd.DeletePrefix(counterKey)
	if err != nil {
		log.Printf("Deleting counters of instance failed: %s", err.Error())
	}

	nameKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagName, configName)
	err = db.Etcd.DeletePrefix(nameKey)
	if err != nil {
		log.Printf("Deleting counters of instance failed: %s", err.Error())
	}

	return nil
}

// Create a version for the configuration. If previous config provided that is also stored
func (c ConfigVersionStore) createConfigVersion(configNew, configPrev Config, action string, resources []KubernetesConfigResource) (uint, error) {

	configName := ""
	if configNew.ConfigName != "" {
		configName = configNew.ConfigName
	} else {
		configName = configPrev.ConfigName
	}

	version, err := c.incrementVersion(configName)

	if err != nil {
		return 0, pkgerrors.Wrap(err, "Get Next Version")
	}
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Retrieving model info")
	}

	versionKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagVersion, configName, strconv.Itoa(int(version)))

	var cs configVersionDBContent
	cs.Action = action
	cs.ConfigNew = configNew
	cs.ConfigPrev = configPrev
	cs.Resources = resources //[]KubernetesConfigResource{}

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
func (c ConfigVersionStore) deleteConfigVersion(configName string) error {

	counter, err := c.getCurrentVersion(configName)

	if err != nil {
		return pkgerrors.Wrap(err, "Get Next Version")
	}
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}
	versionKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagVersion, configName, strconv.Itoa(int(counter)))

	err = db.Etcd.Delete(versionKey)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Config DB Entry")
	}
	err = c.decrementVersion(configName)
	if err != nil {
		return pkgerrors.Wrap(err, "Decrement Version")
	}
	return nil
}

// Read the specified version of the configuration and return its prev and current value.
// Also returns the action for the config version
func (c ConfigVersionStore) getConfigVersion(configName string, version uint) (Config, Config, string, []KubernetesConfigResource, error) {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return Config{}, Config{}, "", []KubernetesConfigResource{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	versionKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagVersion, configName, strconv.Itoa(int(version)))
	configBytes, err := db.Etcd.Get(versionKey)
	if err != nil {
		return Config{}, Config{}, "", []KubernetesConfigResource{}, pkgerrors.Wrap(err, "Get Config Version ")
	}

	if configBytes != nil {
		pr := configVersionDBContent{}
		err = db.DeSerialize(string(configBytes), &pr)
		if err != nil {
			return Config{}, Config{}, "", []KubernetesConfigResource{}, pkgerrors.Wrap(err, "DeSerialize Config Version")
		}
		return pr.ConfigNew, pr.ConfigPrev, pr.Action, pr.Resources, nil
	}
	return Config{}, Config{}, "", []KubernetesConfigResource{}, pkgerrors.Wrap(err, "Invalid data ")
}

// Get the counter for the version
func (c ConfigVersionStore) getCurrentVersion(configName string) (uint, error) {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagCounter, configName)

	value, err := db.Etcd.Get(cfgKey)
	if err != nil {
		if strings.Contains(err.Error(), "Key doesn't exist") == true {
			// Counter not started yet, 0 is invalid value
			return 0, nil
		} else {
			return 0, pkgerrors.Wrap(err, "Get Current Version")
		}
	}

	index, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Invalid counter")
	}
	return uint(index), nil
}

// Update the counter for the version
func (c ConfigVersionStore) updateVersion(configName string, counter uint) error {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagCounter, configName)
	err = db.Etcd.Put(cfgKey, strconv.Itoa(int(counter)))
	if err != nil {
		return pkgerrors.Wrap(err, "Counter DB Entry")
	}
	return nil
}

// Increment the version counter
func (c ConfigVersionStore) incrementVersion(configName string) (uint, error) {

	counter, err := c.getCurrentVersion(configName)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Get Next Counter Value")
	}
	//This is done while Profile lock is taken
	counter++
	err = c.updateVersion(configName, counter)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "Store Next Counter Value")
	}

	return counter, nil
}

// Decrement the version counter
func (c ConfigVersionStore) decrementVersion(configName string) error {

	counter, err := c.getCurrentVersion(configName)
	if err != nil {
		return pkgerrors.Wrap(err, "Get Next Counter Value")
	}
	//This is done while Profile lock is taken
	counter--
	err = c.updateVersion(configName, counter)
	if err != nil {
		return pkgerrors.Wrap(err, "Store Next Counter Value")
	}

	return nil
}

// Get tag list
func (c ConfigVersionStore) getTagList(configName string) ([]string, error) {
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return []string{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	tagKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagName, configName)

	tagKeyList, err := db.Etcd.GetKeys(tagKey)
	if err != nil {
		return []string{}, pkgerrors.Wrap(err, "Config DB Entry Not found")
	}
	result := make([]string, 0)
	for _, tag := range tagKeyList {
		result = append(result, tag[len(tagKey):len(tag)-1])
	}
	return result, nil
}

// Get tag version
func (c ConfigVersionStore) getTagVersion(configName, tagNameValue string) (string, error) {
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Retrieving model info")
	}
	tagKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagName, configName, tagNameValue)

	value, err := db.Etcd.Get(tagKey)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Config DB Entry Not found")
	}
	return string(value), nil
}

// Tag current version
func (c ConfigVersionStore) tagCurrentVersion(configName, tagNameValue string) error {
	currentVersion, err := c.getCurrentVersion(configName)
	if err != nil {
		return pkgerrors.Wrap(err, "Get Current Config Version ")
	}
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(c.instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}
	currentConfig, _, _, _, err := c.getConfigVersion(configName, currentVersion)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving current configuration")
	}

	tagKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagName, configName, tagNameValue)

	err = db.Etcd.Put(tagKey, strconv.Itoa(int(currentVersion)))
	if err != nil {
		return pkgerrors.Wrap(err, "TagIt store DB")
	}

	currentConfig.ConfigTag = tagNameValue
	configValue, err := db.Serialize(currentConfig)
	if err != nil {
		return pkgerrors.Wrap(err, "Serialize Config Value")
	}
	cfgKey := constructKey(rbName, rbVersion, profileName, c.instanceID, tagConfig, configName)
	err = db.Etcd.Put(cfgKey, configValue)
	if err != nil {
		return pkgerrors.Wrap(err, "Config DB Entry")
	}
	return nil
}

// Apply Config
func applyConfig(instanceID string, p Config, pChannel chan configResourceList, action string, resources []KubernetesConfigResource) ([]KubernetesConfigResource, error) {

	rbName, rbVersion, profileName, releaseName, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return []KubernetesConfigResource{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	// Get Template and Resolve the template with values
	crl, err := resolve(rbName, rbVersion, profileName, instanceID, p, releaseName)
	if err != nil {
		return []KubernetesConfigResource{}, pkgerrors.Wrap(err, "Resolve Config")
	}
	var updatedResources (chan []KubernetesConfigResource) = make(chan []KubernetesConfigResource)
	crl.action = action
	crl.resources = resources
	crl.updatedResources = updatedResources
	// Send the configResourceList to the channel. Using select for non-blocking channel
	log.Printf("Before Sent to goroutine %v", crl.profile)
	select {
	case pChannel <- crl:
		log.Printf("Message Sent to goroutine %v", crl.profile)
	default:
	}

	var resultResources []KubernetesConfigResource = <-updatedResources
	return resultResources, nil
}

// Per Profile Go routine to apply the configuration to Cloud Region
func scheduleResources(c chan configResourceList) {
	// Keep thread running
	log.Printf("[scheduleResources]: START thread")
	for {
		data := <-c
		//TODO: ADD Check to see if Application running
		ic := NewInstanceClient()
		resp, err := ic.Find(data.profile.RBName, data.profile.RBVersion, data.profile.ProfileName, nil)
		if (err != nil || len(resp) == 0) && data.action != "STOP" {
			log.Println("Error finding a running instance. Retrying later...")
			data.updatedResources <- []KubernetesConfigResource{}
			continue
		}
		breakThread := false
		switch {
		case data.action == "PUT" || data.action == "POST":
			log.Printf("[scheduleResources]: %v %v %v", data.action, data.profile, data.resourceTemplates)
			resources := []KubernetesConfigResource{}
			for _, inst := range resp {
				k8sClient := KubernetesClient{}
				err = k8sClient.Init(inst.Request.CloudRegion, inst.ID)
				if err != nil {
					log.Printf("Getting CloudRegion Information: %s", err.Error())
					//Move onto the next cloud region
					continue
				}
				for _, res := range data.resourceTemplates {
					var resToCreateOrUpdate = []helm.KubernetesResourceTemplate{res}
					resProceeded, err := k8sClient.createResources(resToCreateOrUpdate, inst.Namespace)
					errCreate := err
					var status string = ""
					if err != nil {
						// assuming - the err represent the resource already exist, so going for update
						resProceeded, err = k8sClient.updateResources(resToCreateOrUpdate, inst.Namespace, false)
						if err != nil {
							log.Printf("Error Creating resources: %s", errCreate.Error())
							log.Printf("Error Updating resources: %s", err.Error())
							break
						} else {
							status = "UPDATED"
						}
					} else {
						status = "CREATED"
					}
					for _, resCreated := range resProceeded {
						resource := KubernetesConfigResource{}
						resource.Resource = resCreated
						resource.Status = status
						resources = append(resources, resource)
					}
				}
			}
			data.updatedResources <- resources
		case data.action == "DELETE":
			log.Printf("[scheduleResources]: DELETE %v %v", data.profile, data.resources)
			for _, inst := range resp {
				k8sClient := KubernetesClient{}
				err = k8sClient.Init(inst.Request.CloudRegion, inst.ID)
				if err != nil {
					log.Printf("Getting CloudRegion Information: %s", err.Error())
					//Move onto the next cloud region
					continue
				}
				var tmpResources []helm.KubernetesResource = []helm.KubernetesResource{}
				for _, res := range data.resources {
					tmpResources = append(tmpResources, res.Resource)
				}
				err = k8sClient.deleteResources(helm.GetReverseK8sResources(tmpResources), inst.Namespace)
				if err != nil {
					log.Printf("Error Deleting resources: %s", err.Error())
					continue
				}
			}
			data.updatedResources <- []KubernetesConfigResource{}

		case data.action == "STOP":
			breakThread = true
		}
		if breakThread {
			break
		}
	}
	log.Printf("[scheduleResources]: STOP thread")
}

//Resolve returns the path where the helm chart merged with
//configuration overrides resides.
var resolve = func(rbName, rbVersion, profileName, instanceId string, p Config, releaseName string) (configResourceList, error) {

	var resTemplates []helm.KubernetesResourceTemplate

	profileClient := rb.NewProfileClient()
	profile, err := profileClient.Get(rbName, rbVersion, profileName)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Reading Profile Data")
	}

	var t rb.ConfigTemplate
	if p.TemplateName == "" && p.ConfigName == "" {
		//for rollback to base definition
		t = rb.ConfigTemplate{}
		t.HasContent = false
	} else {
		t, err = rb.NewConfigTemplateClient().Get(rbName, rbVersion, p.TemplateName)
		if err != nil {
			return configResourceList{}, pkgerrors.Wrap(err, "Getting Template")
		}
		if t.ChartName == "" {
			return configResourceList{}, pkgerrors.New("Invalid template no Chart.yaml file found")
		}
	}

	var def []byte

	if t.HasContent {
		def, err = rb.NewConfigTemplateClient().Download(rbName, rbVersion, p.TemplateName)
		if err != nil {
			return configResourceList{}, pkgerrors.Wrap(err, "Downloading Template")
		}
	} else {
		log.Printf("Using Definition Template as a Configuration Template")
		defClient := rb.NewDefinitionClient()
		definition, err := defClient.Get(rbName, rbVersion)
		if err != nil {
			return configResourceList{}, pkgerrors.Wrap(err, "Get RB Definition")
		}
		def, err = defClient.Download(rbName, rbVersion)
		if err != nil {
			return configResourceList{}, pkgerrors.Wrap(err, "Downloading RB Definition Template")
		}
		t.ChartName = definition.ChartName
	}

	ic := NewInstanceClient()
	instance, err := ic.Get(instanceId)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Getting Instance")
	}

	var finalReleaseName string

	if releaseName == "" {
		finalReleaseName = profile.ReleaseName
	} else {
		finalReleaseName = releaseName
	}

	helmClient := helm.NewTemplateClient(profile.KubernetesVersion,
		profile.Namespace,
		finalReleaseName)

	//copy values from the instance
	valuesArr := []string{}
	if instance.Request.OverrideValues != nil {
		for k, v := range instance.Request.OverrideValues {
			valuesArr = append(valuesArr, k+"="+v)
		}
	}
	valuesArr = append(valuesArr, "k8s-rb-instance-id="+instanceId)
	rawValues, err := helmClient.ProcessValues([]string{}, valuesArr)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Processing values")
	}

	if p.Values != nil {
		for k, v := range p.Values {
			rawValues[k] = v
		}
	}
	//Create a temp file in the system temp folder for values input
	b, err := json.Marshal(rawValues)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Error Marshalling config data")
	}
	data, err := yaml.JSONToYAML(b)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "JSON to YAML")
	}

	outputfile, err := ioutil.TempFile("", "helm-config-values-")
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Got error creating temp file")
	}
	_, err = outputfile.Write([]byte(data))
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Got error writting temp file")
	}
	defer outputfile.Close()

	//Download and process the profile first
	//If everything seems okay, then download the config templates
	prYamlClient, err := profileClient.GetYamlClient(rbName, rbVersion, profileName)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Processing Profile Manifest")
	}

	chartBasePath, err := rb.ExtractTarBall(bytes.NewBuffer(def))
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Extracting Template")
	}

	chartPath := filepath.Join(chartBasePath, t.ChartName)
	resTemplates, crdList, _, err := helmClient.GenerateKubernetesArtifacts(chartPath,
		[]string{prYamlClient.GetValues(), outputfile.Name()},
		nil)
	if err != nil {
		return configResourceList{}, pkgerrors.Wrap(err, "Generate final k8s yaml")
	}
	for _, tmp := range resTemplates {
		crdList = append(crdList, tmp)
	}

	crl := configResourceList{
		resourceTemplates: crdList,
		profile:           profile,
	}

	return crl, nil
}

// Get the Mutex for the Profile
func getProfileData(key string) (*sync.Mutex, chan configResourceList) {
	profileData.Lock()
	defer profileData.Unlock()
	_, ok := profileData.profileLockMap[key]
	if !ok {
		profileData.profileLockMap[key] = &sync.Mutex{}
	}
	_, ok = profileData.resourceChannel[key]
	if !ok {
		profileData.resourceChannel[key] = make(chan configResourceList)
		go scheduleResources(profileData.resourceChannel[key])
		time.Sleep(time.Second * 5)
	}
	return profileData.profileLockMap[key], profileData.resourceChannel[key]
}

func removeProfileData(key string) {
	profileData.Lock()
	defer profileData.Unlock()
	_, ok := profileData.profileLockMap[key]
	if ok {
		delete(profileData.profileLockMap, key)
	}
	_, ok = profileData.resourceChannel[key]
	if ok {
		log.Printf("Stop config thread for %s", key)
		crl := configResourceList{
			action: "STOP",
		}
		profileData.resourceChannel[key] <- crl
		delete(profileData.resourceChannel, key)
	}
}
