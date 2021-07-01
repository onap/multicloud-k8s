/*
Copyright © 2021 Nokia Bell Labs
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

package app

import (
	"fmt"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"helm.sh/helm/v3/pkg/release"
	"log"
	"strings"
	"time"
)

// Timeout used when deleting resources with a hook-delete-policy.
const defaultHookDeleteTimeoutInSeconds = int64(60)

// HookClient implements the Helm Hook interface
type HookClient struct {
	kubeNameSpace 	string
	id     			string
	dbStoreName		string
	dbTagInst		string
}

type MultiCloudHook struct{
	release.Hook
	Group   string
	Version string
}

// NewHookClient returns a new instance of HookClient
func NewHookClient(namespace, id, dbStoreName, dbTagInst string) *HookClient {
	return &HookClient{
		kubeNameSpace: namespace,
		id: id,
		dbStoreName: dbStoreName,
		dbTagInst: dbTagInst,
	}
}

func (hc *HookClient) getHookByEvent(hs []*helm.Hook, hook release.HookEvent) []*helm.Hook {
	hooks := []*helm.Hook{}
	for _, h := range hs {
		for _, e := range h.Hook.Events {
			if e == hook {
				hooks = append(hooks, h)
			}
		}
	}
	return hooks
}

// Mimic function ExecHook in helm/pkg/tiller/release_server.go
func (hc *HookClient) ExecHook(
	k8sClient KubernetesClient,
	hs []*helm.Hook,
	hook release.HookEvent,
	timeout int64,
	startIndex int,
	dbData *InstanceDbData) (error){
	executingHooks := hc.getHookByEvent(hs, hook)
	key := InstanceKey{
		ID: hc.id,
	}
	log.Printf("Executing %d %s hook(s) for instance %s", len(executingHooks), hook, hc.id)
	executingHooks = sortByHookWeight(executingHooks)

	for index, h := range executingHooks {
		if index < startIndex {
			continue
		}
		// Set default delete policy to before-hook-creation
		if h.Hook.DeletePolicies == nil || len(h.Hook.DeletePolicies) == 0 {
			h.Hook.DeletePolicies = []release.HookDeletePolicy{release.HookBeforeHookCreation}
		}
		if err := hc.deleteHookByPolicy(h, release.HookBeforeHookCreation, k8sClient); err != nil {
			return err
		}
		//update DB here before the creation of the hook, if the plugin quits
		//-> when it comes back, it will continue from next hook and consider that this one is done
		if dbData != nil {
			dbData.HookProgress = fmt.Sprintf("%d/%d", index + 1, len(executingHooks))
			err := db.DBconn.Update(hc.dbStoreName, key, hc.dbTagInst, dbData)
			if err != nil {
				return err
			}
		}
		log.Printf("  Instance: %s, Creating %s hook %s, index %d", hc.id, hook, h.Hook.Name, index)
		resTempl := helm.KubernetesResourceTemplate{
			GVK:      h.KRT.GVK,
			FilePath: h.KRT.FilePath,
		}
		createdHook, err := k8sClient.CreateKind(resTempl, hc.kubeNameSpace)
		if  err != nil {
			log.Printf("  Instance: %s, Warning: %s hook %s, filePath: %s, error: %s", hc.id, hook, h.Hook.Name, h.KRT.FilePath, err)
			hc.deleteHookByPolicy(h, release.HookFailed, k8sClient)
			return err
		}
		if hook != "crd-install" {
			//timeout <= 0 -> do not wait
			if timeout > 0 {
				// Watch hook resources until they are completed
				err = k8sClient.WatchHookUntilReady(time.Duration(timeout)*time.Second, hc.kubeNameSpace, createdHook)
				if err != nil {
					// If a hook is failed, check the annotation of the hook to determine whether the hook should be deleted
					// under failed condition. If so, then clear the corresponding resource object in the hook
					if err := hc.deleteHookByPolicy(h, release.HookFailed, k8sClient); err != nil {
						return err
					}
					return err
				}
			}
		} else {
			//Do not handle CRD Hooks
		}
	}

	for _, h := range executingHooks {
		if err := hc.deleteHookByPolicy(h, release.HookSucceeded, k8sClient); err != nil {
			log.Printf("  Instance: %s, Warning: Error deleting %s hook %s based on delete policy, continue", hc.id, hook, h.Hook.Name)
			return err
		}
	}
	log.Printf("%d %s hook(s) complete for release %s", len(executingHooks), hook, hc.id)
	return nil
}

func (hc *HookClient) deleteHookByPolicy(h *helm.Hook, policy release.HookDeletePolicy, k8sClient KubernetesClient) error {
	rss := helm.KubernetesResource{
		GVK:  h.KRT.GVK,
		Name: h.Hook.Name,
	}
	if hookHasDeletePolicy(h, policy) {
		log.Printf("  Instance: %s, Deleting hook %s due to %q policy", hc.id, h.Hook.Name, policy)
		if errHookDelete := k8sClient.deleteResources(append([]helm.KubernetesResource{}, rss), hc.kubeNameSpace); errHookDelete != nil {
			if strings.Contains(errHookDelete.Error(), "not found") {
				return nil
			} else {
				log.Printf("  Instance: %s, Warning: hook %s, filePath %s could not be deleted: %s", hc.id, h.Hook.Name, h.KRT.FilePath ,errHookDelete)
				return errHookDelete
			}
		} else {
			//Verify that the rss is deleted
			isDeleted := false
			for !isDeleted {
				log.Printf("  Instance: %s, Waiting on deleting hook %s for release %s due to %q policy", hc.id, h.Hook.Name, hc.id, policy)
				if _, err := k8sClient.GetResourceStatus(rss, hc.kubeNameSpace); err != nil {
					if strings.Contains(err.Error(), "not found") {
						log.Printf("  Instance: %s, Deleted hook %s for release %s due to %q policy", hc.id, h.Hook.Name, hc.id, policy)
						return nil
					} else {
						isDeleted = true
					}
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
	return nil
}

// hookHasDeletePolicy determines whether the defined hook deletion policy matches the hook deletion polices
// supported by helm. If so, mark the hook as one should be deleted.
func hookHasDeletePolicy(h *helm.Hook, policy release.HookDeletePolicy) bool {
	for _, v := range h.Hook.DeletePolicies {
		if policy == v {
			return true
		}
	}
	return false
}