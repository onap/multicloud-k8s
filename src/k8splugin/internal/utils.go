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

package utils

import (
	"encoding/json"
	"io/ioutil"
	"k8splugin/internal/db"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// LoadedPlugins stores references to the stored plugins
var LoadedPlugins = map[string]*plugin.Plugin{}

const ResourcesListLimit = 10

// ResourceData stores all supported Kubernetes plugin types
type ResourceData struct {
	YamlFilePath string
	Namespace    string
	VnfId        string
}

// DecodeYAML reads a YAMl file to extract the Kubernetes object definition
var DecodeYAML = func(path string, into runtime.Object) (runtime.Object, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + path + " not found")
		} else {
			return nil, pkgerrors.Wrap(err, "Stat file error")
		}
	}

	log.Println("Reading YAML file")
	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read YAML file error")
	}

	log.Println("Decoding deployment YAML")
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

// CheckEnvVariables checks for required Environment variables
func CheckEnvVariables() error {
	envList := []string{"CSAR_DIR", "KUBE_CONFIG_DIR", "PLUGINS_DIR",
		"DATABASE_TYPE", "DATABASE_IP", "OVN_CENTRAL_ADDRESS",
		"AAI_SERVICE_URL", "AAI_USERNAME", "AAI_PASSWORD"}
	for _, env := range envList {
		if _, ok := os.LookupEnv(env); !ok {
			return pkgerrors.New("environment variable " + env + " not set")
		}
	}

	return nil
}

// CheckDatabaseConnection checks if the database is up and running and
// plugin can talk to it
func CheckDatabaseConnection() error {
	err := db.CreateDBClient(os.Getenv("DATABASE_TYPE"))
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = db.DBconn.HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}

// LoadPlugins loads all the compiled .so plugins
func LoadPlugins() error {
	pluginsDir := os.Getenv("PLUGINS_DIR")
	err := filepath.Walk(pluginsDir,
		func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, ".so") {
				p, err := plugin.Open(path)
				if err != nil {
					return pkgerrors.Cause(err)
				}
				LoadedPlugins[info.Name()[:len(info.Name())-3]] = p
			}
			return err
		})
	if err != nil {
		return err
	}

	return nil
}

// CheckInitialSettings is used to check initial settings required to start api
func CheckInitialSettings() error {
	err := CheckEnvVariables()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = CheckDatabaseConnection()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = LoadPlugins()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}

// Vim contains information about the Cloud Region registered in ESR
type Vim struct {
	CloudOwner         string `json:"cloud-owner"`
	CloudRegionID      string `json:"cloud-region-id"`
	CloudType          string `json:"cloud-type"`
	OwnerDefinedType   string `json:"owner-defined-type"`
	CloudRegionVersion string `json:"cloud-region-version"`
	CloudZone          string `json:"cloud-zone"`
}

// VimAuth contains the authentication data to connect to a remote Cloud registered in ESR
type VimAuth struct {
	ESRSystemInfoID string `json:"esr-system-info-id"`
	ServiceURL      string `json:"service-url"`
	Username        string `json:"user-name"`
	Password        string `json:"password"`
	SystemType      string `json:"system-type"`
	SslInsecure     bool   `json:"ssl-insecure"`
	CloudDomain     string `json:"cloud-domain"`
	ResourceVersion string `json:"resource-version"`
}

// GetESRInfo retrieves the Cloud information stored from the ESR GUI portal
func GetESRInfo(cloudOwner, cloudRegionID string) (*Vim, error) {
	if len(cloudOwner) == 0 {
		return nil, pkgerrors.New("cloudOwner empty value")
	}
	if len(cloudRegionID) == 0 {
		return nil, pkgerrors.New("cloudRegionID empty value")
	}
	body, err := doAAIRequest("GET",
		"/cloud-infrastructure/cloud-regions/cloud-region/"+cloudOwner+"/"+cloudRegionID, "")
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The request to retrieve the ESR information failed")
	}

	var vim *Vim
	err = json.Unmarshal([]byte(body), &vim)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The AAI body response is a invalid VIM information")
	}

	return vim, nil
}

// GetESRAuthInfo retrieves the Cloud authentication information stored from the ESR GUI portal
// to connect to a remote Cloud
func GetESRAuthInfo(cloudOwner, cloudRegionID string) (*VimAuth, error) {
	if len(cloudOwner) == 0 {
		return nil, pkgerrors.New("cloudOwner empty value")
	}
	if len(cloudRegionID) == 0 {
		return nil, pkgerrors.New("cloudRegionID empty value")
	}
	body, err := doAAIRequest("GET",
		"/cloud-infrastructure/cloud-regions/cloud-region/"+cloudOwner+"/"+cloudRegionID+
			"/esr-system-info-list/esr-system-info/", "")
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The request to retrieve the ESR Authentication information failed")
	}

	var auth *VimAuth
	err = json.Unmarshal([]byte(body), &auth)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The AAI body response is a invalid VIM authentication data")
	}

	return auth, nil
}

func doAAIRequest(httpMethod, path, payload string) ([]byte, error) {
	aaiURL, err := url.Parse(os.Getenv("AAI_SERVICE_URL"))
	if err != nil && aaiURL.Scheme != "http" && aaiURL.Scheme != "https" {
		return nil, pkgerrors.New("The AAI_SERVICE_URL value is invalid")
	}
	aaiUser := os.Getenv("AAI_USERNAME")
	aaiPassword := os.Getenv("AAI_PASSWORD")
	aaiSchemaVersion := os.Getenv("AAI_SCHEMA_VERSION")
	if len(aaiSchemaVersion) == 0 {
		aaiSchemaVersion = "v14"
	}
	aaiURL.Path = "/aai/" + aaiSchemaVersion + "/" + path

	request, err := http.NewRequest(httpMethod, aaiURL.String(), strings.NewReader(payload))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The AAI request can't be created")
	}
	request.SetBasicAuth(aaiUser, aaiPassword)
	headers := map[string]string{
		"Accept":          "application/json",
		"Content-Type":    "application/json",
		"X-TransactionId": "testaai",
		"X-FromAppId":     "AAI",
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "There is a problem during the execution of the AAI request")
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "The AAI body response can't be read")
	}

	return body, nil
}
