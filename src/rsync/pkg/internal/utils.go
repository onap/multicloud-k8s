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

package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// DecodeYAMLFile reads a YAMl file to extract the Kubernetes object definition
func DecodeYAMLFile(path string, into runtime.Object) (runtime.Object, error) {
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

// DecodeYAMLData reads a string to extract the Kubernetes object definition
func DecodeYAMLData(data string, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(data), nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
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

// TagPodsIfPresent finds the PodTemplateSpec from any workload
// object that contains it and changes the spec to include the tag label
func TagPodsIfPresent(unstruct *unstructured.Unstructured, tag string) {

	spec, ok := unstruct.Object["spec"].(map[string]interface{})
	if !ok {
		log.Println("Error converting spec to map")
		return
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		//log.Println("Error converting template to map")
		return
	}
	log.Println("Apply label in template")
	//Attempt to convert the template to a podtemplatespec.
	//This is to check if we have any pods being created.
	podTemplateSpec := &corev1.PodTemplateSpec{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(template, podTemplateSpec)
	if err != nil {
		log.Println("Did not find a podTemplateSpec: " + err.Error())
		return
	}

	labels := podTemplateSpec.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["emco/deployment-id"] = tag
	podTemplateSpec.SetLabels(labels)

	updatedTemplate, err := runtime.DefaultUnstructuredConverter.ToUnstructured(podTemplateSpec)

	//Set the label
	spec["template"] = updatedTemplate
}
