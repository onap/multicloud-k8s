/*
 * Copyright © 2020 Samsung Electronics
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

package plugin

import (
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
)

func TestTagPodsIfPresent(t *testing.T) {

  testCases := []struct{
    testName                string
    inputUnstructSrc        string
    valueToTag              string
    shouldFailBeforeCheck   bool //This flag provides information if .spec.template.metadata.labels path should be reachable or not
  }{
    {
      testName:               "Resource with no child PodTemplateSpec",
      inputUnstructSrc:       "../../mock_files/mock_yamls/configmap.yaml",
      valueToTag:             "test1",
      shouldFailBeforeCheck:  true,
    },
    {
      testName:               "Deployment with PodTemplateSpec",
      inputUnstructSrc:       "../../mock_files/mock_yamls/deployment.yaml",
      valueToTag:             "test2",
      shouldFailBeforeCheck:  false,
    },
  }

  for _, testCase := range testCases {
    t.Run(testCase.testName, func(t *testing.T){
      holderUnstr := new(unstructured.Unstructured)
      _, err := utils.DecodeYAML(testCase.inputUnstructSrc, holderUnstr)
      if err != nil {
        t.Fatal("Couldn't decode Yaml:", err)
      }
      TagPodsIfPresent(holderUnstr, testCase.valueToTag)
      t.Log(holderUnstr)
      var labelsFinder map[string]interface{} = holderUnstr.Object
      var ok bool
      for _, key := range []string{"spec", "template", "metadata", "labels"} {
        labelsFinder, ok = labelsFinder[key].(map[string]interface{})
        if !ok {
            if testCase.shouldFailBeforeCheck {
              return
            } else {
              t.Fatalf("Error converting %s to map", key)
            }
        }
      }
      if testCase.shouldFailBeforeCheck {
        t.Fatal("Error, nested element '.spec.template.metadata.labels' shouldn't be reachable")
      }
      label, ok := labelsFinder[config.GetConfiguration().KubernetesLabelName].(string)
      if !ok {
        t.Fatalf("Error extracting string label '%s'", config.GetConfiguration().KubernetesLabelName)
      }
      if label != testCase.valueToTag {
        t.Fatalf("Error, expected label '%s' but received '%s'", testCase.valueToTag, label)
      }
    })
  }
}
