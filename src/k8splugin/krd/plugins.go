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

package krd

import (
	"io/ioutil"
	"log"
	"os"
	"plugin"

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

// Initialization Data required by some Plugins
type InitData struct {
	// TODO: Placeholder field. Fields needs to be updated during integration with AA&I
	EsrData string
}

// DecodeYAML reads a YAMl file to extract the Kubernetes object definition
var DecodeYAML = func(path string) (runtime.Object, error) {
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
	obj, _, err := decode(rawBytes, nil, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}
