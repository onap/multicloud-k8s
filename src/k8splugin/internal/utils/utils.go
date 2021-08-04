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
	"log"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"

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
func DecodeYAML(path string, into runtime.Object) (runtime.Object, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + path + " not found")
		} else {
			return nil, pkgerrors.Wrap(err, "Stat file error")
		}
	}

	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read YAML file error")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

// CheckDatabaseConnection checks if the database is up and running and
// plugin can talk to it
func CheckDatabaseConnection() error {
	err := db.CreateDBClient(config.GetConfiguration().DatabaseType)
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = db.DBconn.HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	// TODO Convert these to configuration files instead of environment variables.
	c := db.EtcdConfig{
		Endpoint: config.GetConfiguration().EtcdIP,
		CertFile: config.GetConfiguration().EtcdCert,
		KeyFile:  config.GetConfiguration().EtcdKey,
		CAFile:   config.GetConfiguration().EtcdCAFile,
	}
	err = db.NewEtcdClient(nil, c)
	if err != nil {
		log.Printf("Etcd Client Initialization failed with error: %s", err.Error())
	}
	return nil
}

// LoadPlugins loads all the compiled .so plugins
func LoadPlugins() error {
	pluginsDir := config.GetConfiguration().PluginDir
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
	err := CheckDatabaseConnection()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = LoadPlugins()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}

//EnsureDirectory makes sure that the directories specified in the path exist
//If not, it will create them, if possible.
func EnsureDirectory(f string) error {
	base := path.Dir(f)
	_, err := os.Stat(base)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(base, 0755)
}
