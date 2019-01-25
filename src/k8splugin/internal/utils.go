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
	"io/ioutil"
	"k8splugin/internal/db"
	"log"
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
		"DATABASE_TYPE", "DATABASE_IP", "OVN_CENTRAL_ADDRESS"}
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
