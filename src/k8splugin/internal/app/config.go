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
	"strconv"
	"strings"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"

	pkgerrors "github.com/pkg/errors"
)

// Config contains the parameters needed for configuration
type Config struct {
	ConfigName   string                 `json:"config-name"`
	TemplateName string                 `json:"template-name"`
	Description  string                 `json:"description"`
	Values       map[string]interface{} `json:"values"`
}

//ConfigResult output for Create, Update and delete
type ConfigResult struct {
	InstanceName      string `json:"instance-id"`
	DefinitionName    string `json:"rb-name"`
	DefinitionVersion string `json:"rb-version"`
	ProfileName       string `json:"profile-name"`
	ConfigName        string `json:"config-name"`
	TemplateName      string `json:"template-name"`
	ConfigVersion     uint   `json:"config-verion"`
}

//ConfigRollback input
type ConfigRollback struct {
	AnyOf struct {
		ConfigVersion string `json:"config-version,omitempty"`
		ConfigTag     string `json:"config-tag,omitempty"`
	} `json:"anyOf"`
}

//ConfigTagit for Tagging configurations
type ConfigTagit struct {
	TagName string `json:"tag-name"`
}

// ConfigManager is an interface exposes the config functionality
type ConfigManager interface {
	Create(instanceID string, p Config) (ConfigResult, error)
	Get(instanceID, configName string) (Config, error)
	Help() map[string]string
	Update(instanceID, configName string, p Config) (ConfigResult, error)
	Delete(instanceID, configName string) (ConfigResult, error)
	Rollback(instanceID string, p ConfigRollback) error
	Tagit(instanceID string, p ConfigTagit) error
}

// ConfigClient implements the ConfigManager
// It will also be used to maintain some localized state
type ConfigClient struct {
	tagTag string
}

// NewConfigClient returns an instance of the ConfigClient
// which implements the ConfigManager
func NewConfigClient() *ConfigClient {
	return &ConfigClient{
		tagTag: "tag",
	}
}

// Help returns some information on how to create the content
// for the config in the form of html formatted page
func (v *ConfigClient) Help() map[string]string {
	ret := make(map[string]string)

	return ret
}

// Create an entry for the config in the database
func (v *ConfigClient) Create(instanceID string, p Config) (ConfigResult, error) {

	// Check required fields
	if p.ConfigName == "" || p.TemplateName == "" || len(p.Values) == 0 {
		return ConfigResult{}, pkgerrors.New("Incomplete Configuration Provided")
	}
	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cs := ConfigStore{
		instanceID: instanceID,
		configName: p.ConfigName,
	}
	_, err = cs.getConfig()
	if err == nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Error - Config exists")
	} else {
		if strings.Contains(err.Error(), "Key doesn't exist") == false {
			return ConfigResult{}, pkgerrors.Wrap(err, "Create Error")
		}
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	err = applyConfig(instanceID, p, profileChannel, "POST")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config failed")
	}
	// Create Config DB Entry
	err = cs.createConfig(p)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Config DB Entry")
	}
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	version, err := cvs.createConfigVersion(p, Config{}, "POST")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Config Version DB Entry")
	}

	// Create Result structure
	cfgRes := ConfigResult{
		InstanceName:      instanceID,
		DefinitionName:    rbName,
		DefinitionVersion: rbVersion,
		ProfileName:       profileName,
		ConfigName:        p.ConfigName,
		TemplateName:      p.TemplateName,
		ConfigVersion:     version,
	}
	return cfgRes, nil
}

// Update an entry for the config in the database
func (v *ConfigClient) Update(instanceID, configName string, p Config) (ConfigResult, error) {

	// Check required fields
	if len(p.Values) == 0 {
		return ConfigResult{}, pkgerrors.New("Incomplete Configuration Provided")
	}
	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	_, err = cs.getConfig()
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Error - Config doesn't exist")
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	err = applyConfig(instanceID, p, profileChannel, "PUT")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
	}
	// Update Config DB Entry
	configPrev, err := cs.updateConfig(p)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Config DB Entry")
	}
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	version, err := cvs.createConfigVersion(p, configPrev, "PUT")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Config Version DB Entry")
	}

	// Create Result structure
	cfgRes := ConfigResult{
		InstanceName:      instanceID,
		DefinitionName:    rbName,
		DefinitionVersion: rbVersion,
		ProfileName:       profileName,
		ConfigName:        p.ConfigName,
		TemplateName:      p.TemplateName,
		ConfigVersion:     version,
	}
	return cfgRes, nil
}

// Get config entry in the database
func (v *ConfigClient) Get(instanceID, configName string) (Config, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()
	// Read Config DB
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	cfg, err := cs.getConfig()
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config DB Entry")
	}
	return cfg, nil
}

// Delete the Config from database
func (v *ConfigClient) Delete(instanceID, configName string) (ConfigResult, error) {

	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	p, err := cs.getConfig()
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Error - Config doesn't exist")
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	err = applyConfig(instanceID, p, profileChannel, "DELETE")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
	}
	// Delete Config from DB
	configPrev, err := cs.deleteConfig()
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Delete Config DB Entry")
	}
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	version, err := cvs.createConfigVersion(Config{}, configPrev, "DELETE")
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Delete Config Version DB Entry")
	}

	// Create Result structure
	cfgRes := ConfigResult{
		InstanceName:      instanceID,
		DefinitionName:    rbName,
		DefinitionVersion: rbVersion,
		ProfileName:       profileName,
		ConfigName:        configName,
		TemplateName:      configPrev.TemplateName,
		ConfigVersion:     version,
	}
	return cfgRes, nil
}

// Rollback starts from current version and rollbacks to the version desired
func (v *ConfigClient) Rollback(instanceID string, rback ConfigRollback) error {

	var reqVersion string
	var err error

	if rback.AnyOf.ConfigTag != "" {
		reqVersion, err = v.GetTagVersion(instanceID, rback.AnyOf.ConfigTag)
		if err != nil {
			return pkgerrors.Wrap(err, "Rollback Invalid tag")
		}
	} else if rback.AnyOf.ConfigVersion != "" {
		reqVersion = rback.AnyOf.ConfigVersion
	} else {
		return pkgerrors.Errorf("No valid Index for Rollback")
	}

	index, err := strconv.Atoi(reqVersion)
	if err != nil {
		return pkgerrors.Wrap(err, "Rollback Invalid Index")
	}
	rollbackIndex := uint(index)

	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	currentVersion, err := cvs.getCurrentVersion()
	if err != nil {
		return pkgerrors.Wrap(err, "Rollback Get Current Config Version ")
	}

	if rollbackIndex < 1 && rollbackIndex >= currentVersion {
		return pkgerrors.Wrap(err, "Rollback Invalid Config Version")
	}

	//Rollback all the intermettinent configurations
	for i := currentVersion; i > rollbackIndex; i-- {
		configNew, configPrev, action, err := cvs.getConfigVersion(i)
		if err != nil {
			return pkgerrors.Wrap(err, "Rollback Get Config Version")
		}
		cs := ConfigStore{
			instanceID: instanceID,
			configName: configNew.ConfigName,
		}
		if action == "PUT" {
			// PUT is proceeded by PUT or POST
			err = applyConfig(instanceID, configPrev, profileChannel, "PUT")
			if err != nil {
				return pkgerrors.Wrap(err, "Apply Config  failed")
			}
			_, err = cs.updateConfig(configPrev)
			if err != nil {
				return pkgerrors.Wrap(err, "Update Config DB Entry")
			}
		} else if action == "POST" {
			// POST is always preceeded by Config not existing
			err = applyConfig(instanceID, configNew, profileChannel, "DELETE")
			if err != nil {
				return pkgerrors.Wrap(err, "Delete Config  failed")
			}
			_, err = cs.deleteConfig()
			if err != nil {
				return pkgerrors.Wrap(err, "Delete Config DB Entry")
			}
		} else if action == "DELETE" {
			// DELETE is proceeded by PUT or POST
			err = applyConfig(instanceID, configPrev, profileChannel, "PUT")
			if err != nil {
				return pkgerrors.Wrap(err, "Delete Config  failed")
			}
			_, err = cs.updateConfig(configPrev)
			if err != nil {
				return pkgerrors.Wrap(err, "Update Config DB Entry")
			}
		}
	}
	for i := currentVersion; i > rollbackIndex; i-- {
		// Delete rolled back items
		err = cvs.deleteConfigVersion()
		if err != nil {
			return pkgerrors.Wrap(err, "Delete Config Version ")
		}
	}
	return nil
}

// Tagit tags the current version with the tag provided
func (v *ConfigClient) Tagit(instanceID string, tag ConfigTagit) error {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving model info")
	}
	lock, _ := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	currentVersion, err := cvs.getCurrentVersion()
	if err != nil {
		return pkgerrors.Wrap(err, "Get Current Config Version ")
	}
	tagKey := constructKey(rbName, rbVersion, profileName, instanceID, v.tagTag, tag.TagName)

	err = db.Etcd.Put(tagKey, strconv.Itoa(int(currentVersion)))
	if err != nil {
		return pkgerrors.Wrap(err, "TagIt store DB")
	}
	return nil
}

// GetTagVersion returns the version associated with the tag
func (v *ConfigClient) GetTagVersion(instanceID, tagName string) (string, error) {

	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(instanceID)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Retrieving model info")
	}
	tagKey := constructKey(rbName, rbVersion, profileName, instanceID, v.tagTag, tagName)

	value, err := db.Etcd.Get(tagKey)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Config DB Entry Not found")
	}
	return string(value), nil
}

// ApplyAllConfig starts from first configuration version and applies all versions in sequence
func (v *ConfigClient) ApplyAllConfig(instanceID string) error {

	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
	}
	currentVersion, err := cvs.getCurrentVersion()
	if err != nil {
		return pkgerrors.Wrap(err, "Get Current Config Version ")
	}
	if currentVersion < 1 {
		return pkgerrors.Wrap(err, "No Config Version to Apply")
	}
	//Apply all configurations
	var i uint
	for i = 1; i <= currentVersion; i++ {
		configNew, _, action, err := cvs.getConfigVersion(i)
		if err != nil {
			return pkgerrors.Wrap(err, "Get Config Version")
		}
		err = applyConfig(instanceID, configNew, profileChannel, action)
		if err != nil {
			return pkgerrors.Wrap(err, "Apply Config  failed")
		}
	}
	return nil
}
