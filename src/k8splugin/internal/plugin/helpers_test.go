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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
)

func TestTagPodsIfPresent(t *testing.T) {

  testCases := []struct{
    testName                string
    inputUnstructSrc        string
    valueToTag              string
    isValidOutput           func(*unstructured.Unstructured) bool
    shouldFailBeforeCheck   bool //This flag provides information if .spec.metadata.labels path should be reachable or not
  }{
    {
      testName:               "Resource with no child PodTemplateSpec",
      inputUnstructSrc:       "../../mock_files/mock_yamls/service.yaml",
      isValidOutput:          func(u *unstructured.Unstructured) bool {
        return true
      },
      valueToTag:             "test1",
      shouldFailBeforeCheck:  true,
    },
    {
      testName:               "Deployment with PodTemplateSpec",
      inputUnstructSrc:       "../../mock_files/mock_yamls/deployment.yaml",
      isValidOutput:          func(u *unstructured.Unstructured) bool {return true},
      valueToTag:             "test2",
      shouldFailBeforeCheck:  false,
    },
  }

  for _, testCase := range testCases {
    t.Run(testCase.testName, func(t *testing.T){
      holderUnstr := new(unstructured.Unstructured)
      _, err := utils.DecodeYAML(testCase.inputUnstructSrc, holderUnstr)
      t.Log(holderUnstr)
      if err != nil {
        t.Fatal("Couldn't decode Yaml:", err)
      }
      TagPodsIfPresent(holderUnstr, testCase.valueToTag)
      /*
      spec, ok := holderUnstr.Object["spec"].(map[string]interface{})
      if !ok {
          if testCase.shouldFailBeforeCheck {
            return
          } else {
            t.Fatal("Error converting spec to map")
          }
      }
      template, ok := spec["template"].(map[string]interface{})
      if !ok {
        if testCase.shouldFailBeforeCheck {
          return
        } else {
          t.Fatal("Error converting template to map")
        }
      }
      metadata, ok := template["metadata"].(map[string]interface{})
      if !ok {
        if testCase.shouldFailBeforeCheck {
          return
        } else {
          t.Fatal("Error converting metadata to map")
        }
      }
      labels, ok := metadata["labels"].(map[string]interface{})
      if !ok {
        if testCase.shouldFailBeforeCheck {
          return
        } else {
          t.Fatal("Error converting labels to map")
        }
      }
      */
      defer func(canPanic bool){
        if r := recover(); r != nil {
          if !canPanic {
            t.Fatal("Error, panicked during decoding of unstruct", r)
          }
        }
      }(testCase.shouldFailBeforeCheck)
      labels := holderUnstr.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
      if testCase.shouldFailBeforeCheck {
        t.Fatal("Error, nested element shouldn't be reachable")
      }
      label, ok := labels[config.GetConfiguration().KubernetesLabelName].(string)
      if !ok {
        t.Fatalf("Error extracting string label '%s'", config.GetConfiguration().KubernetesLabelName)
      }
      if label != testCase.valueToTag {
        t.Fatalf("Error, expected label '%s' but received '%s'", testCase.valueToTag, label)
      }
      if !testCase.isValidOutput(holderUnstr) {
        t.Fatal("Returned unstructured doesn't pass validation")
      }
    })
  }
}
