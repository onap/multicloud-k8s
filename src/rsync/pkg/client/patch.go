/*
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
// Based on Code: https://github.com/johandry/klient
package client

import (
	"fmt"
	"os"
	"time"

	"github.com/jonboulle/clockwork"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/resource"
	oapi "k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/openapi"
)

const (
	// overwrite if true, automatically resolve conflicts between the modified and live configuration by using values from the modified configuration
	overwrite = true
	// maxPatchRetry is the maximum number of conflicts retry for during a patch operation before returning failure
	maxPatchRetry = 5
	// backOffPeriod is the period to back off when apply patch results in error.
	backOffPeriod = 1 * time.Second
	// how many times we can retry before back off
	triesBeforeBackOff = 1
	// force if true, immediately remove resources from API and bypass graceful deletion. Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.
	force = false
	// timeout waiting for the resource to be delete if it needs to be recreated
	timeout = 0
)

// patch tries to patch an OpenAPI resource
func patch(info *resource.Info, current runtime.Object) error {
	// From: k8s.io/kubectl/pkg/cmd/apply/apply.go & patcher.go
	modified, err := util.GetModifiedConfiguration(info.Object, true, unstructured.UnstructuredJSONScheme)
	if err != nil {
		return fmt.Errorf("retrieving modified configuration. %s", err)
	}

	metadata, _ := meta.Accessor(current)
	annotationMap := metadata.GetAnnotations()
	if _, ok := annotationMap[corev1.LastAppliedConfigAnnotation]; !ok {
		// TODO: Find what to do with the warnings, they should not be printed
		fmt.Fprintf(os.Stderr, "Warning: apply should be used on resource created by apply")
	}

	patchBytes, patchObject, err := patchSimple(current, modified, info)

	var getErr error
	for i := 1; i <= maxPatchRetry && errors.IsConflict(err); i++ {
		if i > triesBeforeBackOff {
			clockwork.NewRealClock().Sleep(backOffPeriod)
		}
		current, getErr = resource.NewHelper(info.Client, info.Mapping).Get(info.Namespace, info.Name, false)
		if getErr != nil {
			return getErr
		}
		patchBytes, patchObject, err = patchSimple(current, modified, info)
	}
	if err != nil && (errors.IsConflict(err) || errors.IsInvalid(err)) && force {
		patchBytes, patchObject, err = deleteAndCreate(info, patchBytes)
	}

	info.Refresh(patchObject, true)

	return nil
}

func patchSimple(currentObj runtime.Object, modified []byte, info *resource.Info) ([]byte, runtime.Object, error) {
	// Serialize the current configuration of the object from the server.
	current, err := runtime.Encode(unstructured.UnstructuredJSONScheme, currentObj)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing current configuration. %s", err)
	}

	// Retrieve the original configuration of the object from the annotation.
	original, err := util.GetOriginalConfiguration(currentObj)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving original configuration. %s", err)
	}

	var patchType types.PatchType
	var patch []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta
	var schema oapi.Schema

	// Create the versioned struct from the type defined in the restmapping
	// (which is the API version we'll be submitting the patch to)
	versionedObject, err := scheme.Scheme.New(info.Mapping.GroupVersionKind)

	// DEBUG:
	// fmt.Printf("Modified: %v\n", string(modified))
	// fmt.Printf("Current: %v\n", string(current))
	// fmt.Printf("Original: %v\n", string(original))
	// fmt.Printf("versionedObj: %v\n", versionedObject)
	// fmt.Printf("Error: %+v\nIsNotRegisteredError: %t\n", err, runtime.IsNotRegisteredError(err))

	switch {
	case runtime.IsNotRegisteredError(err):
		// fall back to generic JSON merge patch
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			if mergepatch.IsPreconditionFailed(err) {
				return nil, nil, fmt.Errorf("At least one of apiVersion, kind and name was changed")
			}
			return nil, nil, fmt.Errorf("creating patch. %s", err)
		}
	case err != nil:
		return nil, nil, fmt.Errorf("getting instance of versioned object. %s", err)
	case err == nil:
		// Compute a three way strategic merge patch to send to server.
		patchType = types.StrategicMergePatchType

		// Try to use openapi first if the openapi spec is available and can successfully calculate the patch.
		// Otherwise, fall back to baked-in types.
		var openapiSchema openapi.Resources
		if openapiSchema != nil {
			if schema = openapiSchema.LookupResource(info.Mapping.GroupVersionKind); schema != nil {
				lookupPatchMeta = strategicpatch.PatchMetaFromOpenAPI{Schema: schema}
				if openapiPatch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, overwrite); err == nil {
					patchType = types.StrategicMergePatchType
					patch = openapiPatch
					// TODO: In case it's necessary to report warnings
					// } else {
					// 	log.Printf("Warning: error calculating patch from openapi spec: %s", err)
				}
			}
		}

		if patch == nil {
			lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
			if err != nil {
				return nil, nil, fmt.Errorf("creating patch. %s", err)
			}
			patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, overwrite)
			if err != nil {
				return nil, nil, fmt.Errorf("creating patch. %s", err)
			}
		}
	}

	if string(patch) == "{}" {
		return patch, currentObj, nil
	}

	patchedObj, err := resource.NewHelper(info.Client, info.Mapping).Patch(info.Namespace, info.Name, patchType, patch, nil)
	return patch, patchedObj, err
}

func deleteAndCreate(info *resource.Info, modified []byte) ([]byte, runtime.Object, error) {
	delOptions := defaultDeleteOptions()
	if _, err := deleteWithOptions(info, delOptions); err != nil {
		return nil, nil, err
	}

	helper := resource.NewHelper(info.Client, info.Mapping)

	// TODO: make a waiter and use it
	if err := wait.PollImmediate(1*time.Second, time.Duration(timeout), func() (bool, error) {
		if _, err := helper.Get(info.Namespace, info.Name, false); !errors.IsNotFound(err) {
			return false, err
		}
		return true, nil
	}); err != nil {
		return nil, nil, err
	}

	// TODO: Check what GetModifiedConfiguration does, this could be an encode - decode waste of time
	// modified, err := util.GetModifiedConfiguration(info.Object, true, unstructured.UnstructuredJSONScheme)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("retrieving modified configuration. %s", err)
	// }
	versionedObject, _, err := unstructured.UnstructuredJSONScheme.Decode(modified, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	options := metav1.CreateOptions{}
	createdObject, err := helper.Create(info.Namespace, true, versionedObject, &options)
	if err != nil {
		// restore the original object if we fail to create the new one
		// but still propagate and advertise error to user
		recreated, recreateErr := helper.Create(info.Namespace, true, info.Object, &options)
		if recreateErr != nil {
			err = fmt.Errorf("An error occurred force-replacing the existing object with the newly provided one. %v.\n\nAdditionally, an error occurred attempting to restore the original object: %v", err, recreateErr)
		} else {
			createdObject = recreated
		}
	}
	return modified, createdObject, err
}
