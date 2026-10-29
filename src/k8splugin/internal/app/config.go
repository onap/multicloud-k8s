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
	"context"
	"strconv"
	"strings"

	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	pkgerrors "github.com/pkg/errors"
)

// Config contains the parameters needed for configuration
type Config struct {
	ConfigName    string                 `json:"config-name"`
	TemplateName  string                 `json:"template-name"`
	Description   string                 `json:"description"`
	Values        map[string]interface{} `json:"values"`
	ConfigVersion uint                   `json:"config-version"`
	ConfigTag     string                 `json:"config-tag"`
}

// ConfigResult output for Create, Update and delete
type ConfigResult struct {
	InstanceName      string `json:"instance-id"`
	DefinitionName    string `json:"rb-name"`
	DefinitionVersion string `json:"rb-version"`
	ProfileName       string `json:"profile-name"`
	ConfigName        string `json:"config-name"`
	TemplateName      string `json:"template-name"`
	ConfigVersion     uint   `json:"config-version"`
}

// ConfigRollback input
type ConfigRollback struct {
	AnyOf struct {
		ConfigVersion string `json:"config-version,omitempty"`
		ConfigTag     string `json:"config-tag,omitempty"`
	} `json:"anyOf"`
}

// ConfigRollback input
type ConfigTag struct {
	ConfigVersion uint   `json:"config-version"`
	ConfigTag     string `json:"config-tag"`
}

// ConfigTagit for Tagging configurations
type ConfigTagit struct {
	TagName string `json:"tag-name"`
}

// ConfigManager is an interface exposes the config functionality
type ConfigManager interface {
	Create(ctx context.Context, instanceID string, p Config) (ConfigResult, error)
	Get(ctx context.Context, instanceID, configName string) (Config, error)
	GetVersion(ctx context.Context, instanceID, configName, version string) (Config, error)
	GetTag(ctx context.Context, instanceID, configName, tagName string) (Config, error)
	List(ctx context.Context, instanceID string) ([]Config, error)
	VersionList(ctx context.Context, instanceID, configName string) ([]Config, error)
	Help() map[string]string
	Update(ctx context.Context, instanceID, configName string, p Config) (ConfigResult, error)
	Delete(ctx context.Context, instanceID, configName string) (ConfigResult, error)
	DeleteAll(ctx context.Context, instanceID, configName string, deleteConfigOnly bool) error
	Rollback(ctx context.Context, instanceID string, configName string, p ConfigRollback, acceptRevert bool) (ConfigResult, error)
	Cleanup(ctx context.Context, instanceID string) error
	Tagit(ctx context.Context, instanceID string, configName string, p ConfigTagit) (ConfigTag, error)
	TagList(ctx context.Context, instanceID, configName string) ([]ConfigTag, error)
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
func (v *ConfigClient) Create(ctx context.Context, instanceID string, p Config) (ConfigResult, error) {
	log.InfoContext(ctx, "Config Create", log.Fields{"instance": instanceID})
	// Check required fields
	if p.ConfigName == "" || p.TemplateName == "" {
		return ConfigResult{}, pkgerrors.New("Incomplete Configuration Provided")
	}
	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(ctx, instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	cs := ConfigStore{
		instanceID: instanceID,
		configName: p.ConfigName,
	}
	_, err = cs.getConfig(ctx)
	if err == nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Error - Config exists")
	} else {
		if !strings.Contains(err.Error(), "Key doesn't exist") {
			return ConfigResult{}, pkgerrors.Wrap(err, "Create Error")
		}
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	var appliedResources ([]KubernetesConfigResource)
	appliedResources, err = applyConfig(ctx, instanceID, p, profileChannel, "POST", nil)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config failed")
	}
	log.InfoContext(ctx, "POST result", log.Fields{"resources": appliedResources})
	// Create Config DB Entry
	err = cs.createConfig(ctx, p)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Config DB Entry")
	}
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: p.ConfigName,
	}
	version, err := cvs.createConfigVersion(ctx, p, Config{}, "POST", appliedResources)
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
func (v *ConfigClient) Update(ctx context.Context, instanceID, configName string, p Config) (ConfigResult, error) {
	log.InfoContext(ctx, "Config Update", log.Fields{"instance": instanceID, "config": configName})
	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(ctx, instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	_, err = cs.getConfig(ctx)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Error - Config doesn't exist")
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	var appliedResources ([]KubernetesConfigResource)
	appliedResources, err = applyConfig(ctx, instanceID, p, profileChannel, "PUT", nil)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
	}
	log.InfoContext(ctx, "PUT result", log.Fields{"resources": appliedResources})
	// Update Config DB Entry
	configPrev, err := cs.updateConfig(ctx, p)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Config DB Entry")
	}
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	version, err := cvs.createConfigVersion(ctx, p, configPrev, "PUT", appliedResources)
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
func (v *ConfigClient) Get(ctx context.Context, instanceID, configName string) (Config, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()
	// Read Config DB
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	cfg, err := cs.getConfig(ctx)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config DB Entry")
	}

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config Version Entry")
	}
	cfg.ConfigVersion = currentVersion
	return cfg, nil
}

// Get version config entry in the database
func (v *ConfigClient) GetTag(ctx context.Context, instanceID, configName, tagName string) (Config, error) {
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	version, err := cvs.getTagVersion(ctx, configName, tagName)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config Tag Version Entry")
	}
	return v.GetVersion(ctx, instanceID, configName, version)
}

// Get version config entry in the database
func (v *ConfigClient) GetVersion(ctx context.Context, instanceID, configName, version string) (Config, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()
	// Read Config DB
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	cfg, _ := cs.getConfig(ctx)

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	versionInt, err := strconv.ParseUint(version, 0, 32)
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Parsint version string")
	}
	_, _, _, _, err = cvs.getConfigVersion(ctx, configName, uint(versionInt))
	if err != nil {
		return Config{}, pkgerrors.Wrap(err, "Get Config Version Entry")
	}
	cfg.ConfigVersion = uint(versionInt)
	return cfg, nil
}

// List config entry in the database
func (v *ConfigClient) List(ctx context.Context, instanceID string) ([]Config, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()
	// Read Config DB
	cs := ConfigStore{
		instanceID: instanceID,
	}
	cfg, err := cs.getConfigList(ctx)
	result := make([]Config, 0)
	for _, config := range cfg {
		cvs := ConfigVersionStore{
			instanceID: instanceID,
			configName: config.ConfigName,
		}
		currentVersion, err := cvs.getCurrentVersion(ctx, config.ConfigName)
		if err != nil {
			return []Config{}, pkgerrors.Wrap(err, "Get Current Config Version ")
		}
		config.ConfigVersion = currentVersion
		result = append(result, config)
	}
	if err != nil {
		return []Config{}, pkgerrors.Wrap(err, "Get Config DB Entry")
	}
	return result, nil
}

// Version List config entry in the database
func (v *ConfigClient) VersionList(ctx context.Context, instanceID string, configName string) ([]Config, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return []Config{}, pkgerrors.Wrap(err, "Get Current Config Version ")
	}
	//Get all configurations
	var i uint
	cfgList := make([]Config, 0)
	for i = 1; i <= currentVersion; i++ {
		config, _, _, _, err := cvs.getConfigVersion(ctx, configName, i)
		config.ConfigVersion = i
		if err != nil {
			return []Config{}, pkgerrors.Wrap(err, "Get Config Version")
		}
		cfgList = append(cfgList, config)
	}

	return cfgList, nil
}

func (v *ConfigClient) TagList(ctx context.Context, instanceID, configName string) ([]ConfigTag, error) {

	// Acquire per profile Mutex
	lock, _ := getProfileData(instanceID)
	lock.Lock()
	defer lock.Unlock()
	// Read Config DB
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	_, err := cs.getConfig(ctx)
	if err != nil {
		return []ConfigTag{}, pkgerrors.Wrap(err, "Get Config DB Entry")
	}
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}

	tagList, err := cvs.getTagList(ctx, configName)
	if err != nil {
		return []ConfigTag{}, pkgerrors.Wrap(err, "Get Tag list")
	}
	result := make([]ConfigTag, 0)
	for _, tag := range tagList {
		tagData := ConfigTag{}
		version, err := cvs.getTagVersion(ctx, configName, tag)
		if err != nil {
			return []ConfigTag{}, pkgerrors.Wrap(err, "Get Tag version")
		}
		versionInt, err := strconv.ParseUint(version, 0, 32)
		if err != nil {
			return []ConfigTag{}, pkgerrors.Wrap(err, "Parsint version string")
		}
		tagData.ConfigTag = tag
		tagData.ConfigVersion = uint(versionInt)
		result = append(result, tagData)
	}
	return result, nil
}

// Delete the Config from database
func (v *ConfigClient) DeleteAll(ctx context.Context, instanceID, configName string, deleteConfigOnly bool) error {
	log.InfoContext(ctx, "Config Delete All", log.Fields{"instance": instanceID, "config": configName})
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	_, err := cs.getConfig(ctx)
	if err != nil {
		return pkgerrors.Wrap(err, "Update Error - Config doesn't exist")
	}
	// Get Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}

	if !deleteConfigOnly {
		var rollbackConfig = ConfigRollback{}
		rollbackConfig.AnyOf.ConfigVersion = "0"
		_, err = v.Rollback(ctx, instanceID, configName, rollbackConfig, true)
		if err != nil {
			return pkgerrors.Wrap(err, "Rollback to base version")
		}
	}
	// Delete Config from DB
	_, err = cs.deleteConfig(ctx)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Config DB Entry")
	}
	cvs.cleanupIstanceTags(ctx, configName)
	return nil
}

// Apply update with delete operation
func (v *ConfigClient) Delete(ctx context.Context, instanceID, configName string) (ConfigResult, error) {
	log.InfoContext(ctx, "Config Delete", log.Fields{"instance": instanceID, "config": configName})
	// Resolving rbName, Version, etc. not to break response
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(ctx, instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	p, err := cs.getConfig(ctx)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Delete Error - Config doesn't exist")
	}
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()
	// Create Version Entry in DB for Config
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Current version get failed")
	}
	_, _, _, resources, err := cvs.getConfigVersion(ctx, configName, currentVersion)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Config version get failed")
	}

	_, err = applyConfig(ctx, instanceID, p, profileChannel, "DELETE", resources)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
	}
	log.InfoContext(ctx, "DELETE resources", log.Fields{"resources": resources})
	// Update Config from DB
	configPrev, err := cs.updateConfig(ctx, p)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Config DB Entry")
	}
	version, err := cvs.createConfigVersion(ctx, p, configPrev, "DELETE", []KubernetesConfigResource{})
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Create Delete Config Version DB Entry")
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
func (v *ConfigClient) Rollback(ctx context.Context, instanceID string, configName string, rback ConfigRollback, acceptRevert bool) (ConfigResult, error) {
	log.InfoContext(ctx, "Config Rollback", log.Fields{"instance": instanceID, "config": configName})
	var reqVersion string
	var err error
	rbName, rbVersion, profileName, _, err := resolveModelFromInstance(ctx, instanceID)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Retrieving model info")
	}
	if rback.AnyOf.ConfigTag != "" {
		reqVersion, err = v.GetTagVersion(ctx, instanceID, configName, rback.AnyOf.ConfigTag)
		if err != nil {
			return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Invalid tag")
		}
	} else if rback.AnyOf.ConfigVersion != "" {
		reqVersion = rback.AnyOf.ConfigVersion
	} else {
		return ConfigResult{}, pkgerrors.Errorf("No valid Index for Rollback")
	}

	index, err := strconv.Atoi(reqVersion)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Invalid Index")
	}
	rollbackIndex := uint(index)

	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Get Current Config Version ")
	}

	if (rollbackIndex < 1 && !acceptRevert) || rollbackIndex >= currentVersion {
		return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Invalid Config Version")
	}

	if rollbackIndex < 1 && acceptRevert {
		rollbackIndex = 0
	}

	//Rollback all the intermettinent configurations
	for i := currentVersion; i > rollbackIndex; i-- {
		configNew, configPrev, _, resources, err := cvs.getConfigVersion(ctx, configName, i)
		if err != nil {
			return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Get Config Version")
		}
		var prevAction string
		if i == 1 {
			prevAction = "POST"
			configPrev.ConfigName = ""
			configPrev.TemplateName = ""
			configPrev.Values = make(map[string]interface{})
		} else {
			_, _, prevAction, _, err = cvs.getConfigVersion(ctx, configName, i-1)
		}
		log.InfoContext(ctx, "ROLLBACK to version", log.Fields{"version": i - 1})
		if err != nil {
			return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Get Prev Config Version")
		}
		cs := ConfigStore{
			instanceID: instanceID,
			configName: configNew.ConfigName,
		}
		if prevAction != "DELETE" {
			var resourcesToDelete = make([]KubernetesConfigResource, 0)
			for _, res := range resources {
				if res.Status == "CREATED" {
					resourcesToDelete = append(resourcesToDelete, res)
				}
			}
			if len(resourcesToDelete) > 0 {
				_, err := applyConfig(ctx, instanceID, configPrev, profileChannel, "DELETE", resources)
				if err != nil {
					return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
				}
			}
			appliedResources, err := applyConfig(ctx, instanceID, configPrev, profileChannel, prevAction, nil)
			if err != nil {
				return ConfigResult{}, pkgerrors.Wrap(err, "Apply Config  failed")
			}
			log.InfoContext(ctx, "Rollback result", log.Fields{"action": prevAction, "resources": appliedResources})
			_, err = cs.updateConfig(ctx, configPrev)
			if err != nil {
				return ConfigResult{}, pkgerrors.Wrap(err, "Update Config DB Entry")
			}
		} else {
			// POST is always preceeded by Config not existing
			_, err := applyConfig(ctx, instanceID, configPrev, profileChannel, prevAction, resources)
			if err != nil {
				return ConfigResult{}, pkgerrors.Wrap(err, "Delete Config  failed")
			}
			log.InfoContext(ctx, "DELETE resources", log.Fields{"resources": resources})
			_, err = cs.updateConfig(ctx, configPrev)
			if err != nil {
				return ConfigResult{}, pkgerrors.Wrap(err, "Update Config DB Entry")
			}
		}
	}
	if rollbackIndex == 0 {
		//this is used only for delete config and remianing configuration 1 will be removed there
		rollbackIndex = 1
	}
	for i := currentVersion; i > rollbackIndex; i-- {
		// Delete rolled back items
		err = cvs.deleteConfigVersion(ctx, configName)
		if err != nil {
			return ConfigResult{}, pkgerrors.Wrap(err, "Delete Config Version ")
		}
	}
	currentVersion, err = cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Rollback Get Current Config Version ")
	}
	// Check if Config exists
	cs := ConfigStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentConfig, err := cs.getConfig(ctx)
	if err != nil {
		return ConfigResult{}, pkgerrors.Wrap(err, "Update Error - Config doesn't exist")
	}
	// Create Result structure
	cfgRes := ConfigResult{
		InstanceName:      instanceID,
		DefinitionName:    rbName,
		DefinitionVersion: rbVersion,
		ProfileName:       profileName,
		ConfigName:        configName,
		TemplateName:      currentConfig.TemplateName,
		ConfigVersion:     currentVersion,
	}
	return cfgRes, nil
}

// Tagit tags the current version with the tag provided
func (v *ConfigClient) Tagit(ctx context.Context, instanceID string, configName string, tag ConfigTagit) (ConfigTag, error) {
	log.InfoContext(ctx, "Config Tag It", log.Fields{"instance": instanceID, "config": configName})
	lock, _ := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	err := cvs.tagCurrentVersion(ctx, configName, tag.TagName)
	if err != nil {
		return ConfigTag{}, pkgerrors.Wrap(err, "Tag of current version failed")
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return ConfigTag{}, pkgerrors.Wrap(err, "Rollback Get Current Config Version ")
	}

	var tagResult = ConfigTag{}
	tagResult.ConfigVersion = currentVersion
	tagResult.ConfigTag = tag.TagName
	return tagResult, nil
}

// GetTagVersion returns the version associated with the tag
func (v *ConfigClient) GetTagVersion(ctx context.Context, instanceID, configName string, tagName string) (string, error) {
	log.InfoContext(ctx, "Config Get Tag Version", log.Fields{"instance": instanceID, "config": configName})
	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	value, err := cvs.getTagVersion(ctx, configName, tagName)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Tag of current version failed")
	}

	return value, nil
}

// Cleanup version used only when instance is being deleted. We do not pass errors and we try to delete data
func (v *ConfigClient) Cleanup(ctx context.Context, instanceID string) error {
	log.InfoContext(ctx, "Config Cleanup", log.Fields{"instance": instanceID})
	configs, err := v.List(ctx, instanceID)

	if err != nil {
		return pkgerrors.Wrap(err, "Retrieving active config list info")
	}

	for _, config := range configs {
		_, err = v.Delete(ctx, instanceID, config.ConfigName)
		if err != nil {
			log.ErrorContext(ctx, "Config delete failed", log.Fields{"config": config.ConfigName, "error": err})
		}
		err = v.DeleteAll(ctx, instanceID, config.ConfigName, true)
		if err != nil {
			log.ErrorContext(ctx, "Config delete failed", log.Fields{"config": config.ConfigName, "error": err})
		}
	}

	removeProfileData(ctx, instanceID)

	return nil
}

// ApplyAllConfig starts from first configuration version and applies all versions in sequence
func (v *ConfigClient) ApplyAllConfig(ctx context.Context, instanceID string, configName string) error {
	log.InfoContext(ctx, "Config Apply All", log.Fields{"instance": instanceID, "config": configName})
	lock, profileChannel := getProfileData(instanceID)
	// Acquire per profile Mutex
	lock.Lock()
	defer lock.Unlock()

	cvs := ConfigVersionStore{
		instanceID: instanceID,
		configName: configName,
	}
	currentVersion, err := cvs.getCurrentVersion(ctx, configName)
	if err != nil {
		return pkgerrors.Wrap(err, "Get Current Config Version ")
	}
	if currentVersion < 1 {
		return pkgerrors.Wrap(err, "No Config Version to Apply")
	}
	//Apply all configurations
	var i uint
	for i = 1; i <= currentVersion; i++ {
		configNew, _, action, resources, err := cvs.getConfigVersion(ctx, configName, i)
		if err != nil {
			return pkgerrors.Wrap(err, "Get Config Version")
		}
		if action != "DELETE" {
			resources = nil
		}
		var appliedResources ([]KubernetesConfigResource)
		appliedResources, err = applyConfig(ctx, instanceID, configNew, profileChannel, action, resources)
		if err != nil {
			return pkgerrors.Wrap(err, "Apply Config  failed")
		}
		log.InfoContext(ctx, "Apply result", log.Fields{"action": action, "resources": appliedResources})
	}
	return nil
}
