/*
Copyright © 2020 Intel Corp

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
package cmd

import (
	"strconv"
)

// Configurations exported
type EmcoConfigurations struct {
	Orchestrator ControllerConfigurations
	Clm          ControllerConfigurations
	Ncm          ControllerConfigurations
	Dcm          ControllerConfigurations
	Rsync        ControllerConfigurations
	OvnAction    ControllerConfigurations
}

// ControllerConfigurations exported
type ControllerConfigurations struct {
	Port int
	Host string
}

const urlVersion string = "v2"
const urlPrefix string = "http://"
var Configurations EmcoConfigurations

// SetDefaultConfiguration default configuration if t
func SetDefaultConfiguration() {
	Configurations.Orchestrator.Host = "localhost"
	Configurations.Orchestrator.Port = 9015
	Configurations.Clm.Host = "localhost"
	Configurations.Clm.Port = 9061
	Configurations.Ncm.Host = "localhost"
	Configurations.Ncm.Port = 9031
	Configurations.Dcm.Host = "localhost"
	Configurations.Dcm.Port = 0
	Configurations.OvnAction.Host = "localhost"
	Configurations.OvnAction.Port = 9051
}

// GetOrchestratorURL Url for Orchestrator
func GetOrchestratorURL() string {
	if Configurations.Orchestrator.Host == "" || Configurations.Orchestrator.Port == 0 {
		panic("No Orchestrator Information in Config File")
	}
	return urlPrefix + Configurations.Orchestrator.Host + ":" + strconv.Itoa(Configurations.Orchestrator.Port) + "/" + urlVersion
}

// GetClmURL Url for Clm
func GetClmURL() string {
	if Configurations.Clm.Host == "" || Configurations.Clm.Port == 0 {
		panic("No Clm Information in Config File")
	}
	return urlPrefix + Configurations.Clm.Host + ":" + strconv.Itoa(Configurations.Clm.Port) + "/" + urlVersion
}

// GetNcmURL Url for Ncm
func GetNcmURL() string {
	if Configurations.Ncm.Host == "" || Configurations.Ncm.Port == 0 {
		panic("No Ncm Information in Config File")
	}
	return urlPrefix + Configurations.Ncm.Host + ":" + strconv.Itoa(Configurations.Ncm.Port) + "/" + urlVersion
}

// GetDcmURL Url for Dcm
func GetDcmURL() string {
	if Configurations.Dcm.Host == "" || Configurations.Dcm.Port == 0 {
		panic("No Dcm Information in Config File")
	}
	return urlPrefix + Configurations.Dcm.Host + ":" + strconv.Itoa(Configurations.Dcm.Port) + "/" + urlVersion
}

// GetOvnactionURL Url for Ovnaction
func GetOvnactionURL() string {
	if Configurations.OvnAction.Host == "" || Configurations.OvnAction.Port == 0 {
		panic("No Ovn Action Information in Config File")
	}
	return urlPrefix + Configurations.OvnAction.Host + ":" + strconv.Itoa(Configurations.OvnAction.Port) + "/" + urlVersion
}
