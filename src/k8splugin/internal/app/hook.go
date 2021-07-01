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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"helm.sh/helm/v3/pkg/release"
	"log"
	"math"
	"strings"
	"time"
)

// Timeout used when deleting resources with a hook-delete-policy.
const defaultHookDeleteTimeoutInSeconds = int64(60)

// HookClient implements the Helm Hook interface
type HookClient struct {
	kubeNameSpace string
	id     string
}

type MultiCloudHook struct{
	release.Hook
	Group   string
	Version string
}

type result struct {
	hooks   []*MultiCloudHook
	generic []helm.KubernetesResourceTemplate
}

// NewHookClient returns a new instance of HookClient
func NewHookClient(namespace, id string) *HookClient {
	return &HookClient{
		kubeNameSpace: namespace,
		id: id,
	}
}

// Mimic function ExecHook in helm/pkg/tiller/release_server.go
func (hc *HookClient) ExecHook(k8sClient KubernetesClient, hs []*helm.Hook, name string, hook release.HookEvent, timeout int64) ([]helm.KubernetesResource, []*helm.Hook, error){
	createdHooks := []helm.KubernetesResource{}
	executingHooks := []*helm.Hook{}
	for _, h := range hs {
		for _, e := range h.Hook.Events {
			if e == hook {
				executingHooks = append(executingHooks, h)
			}
		}
	}
	log.Printf("Executing %d %s hook(s) for release %s", len(executingHooks), hook, name)
	executingHooks = sortByHookWeight(executingHooks)

	for _, h := range executingHooks {
		if err := hc.deleteHookByPolicy(h, release.HookBeforeHookCreation, name, hook, k8sClient); err != nil {
			return createdHooks, executingHooks, err
		}
		log.Printf("  Release: %s, Creating %s hook %s", name, hook, h.Hook.Name)
		createdHook, err := k8sClient.CreateHookResources(h, hc.kubeNameSpace);
		if  err != nil {
			log.Printf("  Release: %s, Warning: %s hook %s, filePath %s failed: %s", name, hook, h.Hook.Name, h.KRT.FilePath, err)
			hc.deleteHookByPolicy(h, release.HookFailed, name, hook, k8sClient);
			return createdHooks, executingHooks, err
		}
		isReady := false
		isCreated := false
		count := 0
		countMax := int(math.Floor(float64(timeout)/5))
		if hook != "crd-install" {
			for !isReady {
				status, err := k8sClient.GetResourceStatus(*createdHook, hc.kubeNameSpace)
				if err != nil {
					// To avoid the case when the hook is deleted too fast
					if isCreated && strings.Contains(err.Error(), "not found") {
						log.Printf("  Release: %s, Warning: %s hook %s, filePath %s is created and deleted, continue", name, hook, h.Hook.Name, h.KRT.FilePath)
						break
					}
					log.Printf("  Release: %s, Warning: %s hook %s, filePath %s could not complete: %s", name, hook, h.Hook.Name, h.KRT.FilePath, err)
					return createdHooks, executingHooks, err
				}
				isReady, err := IsRssReady(h.Hook.Name ,status, false)
				if isReady {
					log.Printf("  Release: %s, creation complete for %s hook %s", name, h.Hook.Name, hook)
					break
				} else {
					isCreated = true
					log.Printf("  Release: %s, %s hook %s is NOT READY, wait", name, hook, h.Hook.Name)
					time.Sleep(5 * time.Second)
					// A hook cannot be longer than timeout
					if count++; count >= countMax {
						log.Printf("  Release: %s, Warning: %s hook %s cannot complete, consider as failed and continue to next hook", name, hook, h.Hook.Name)
						isReady = true
						if err := hc.deleteHookByPolicy(h, release.HookFailed, name, hook, k8sClient); err != nil {
							return createdHooks, executingHooks, err
						}
						break
					}
				}
			}
			createdHooks = append(createdHooks, helm.KubernetesResource{
				GVK: h.KRT.GVK,
				Name: h.Hook.Name,
			})
		} else {
			//Do not handle CRD Hooks
			createdHooks = append(createdHooks, helm.KubernetesResource{
				GVK: h.KRT.GVK,
				Name: h.Hook.Name,
			})
		}
	}

	log.Printf("%d %s hook(s) complete for release %s", len(executingHooks), hook, name)
	go func() {
		for _, h := range executingHooks {
			if err := hc.deleteHookByPolicy(h, release.HookSucceeded, name, hook, k8sClient); err != nil {
				log.Printf("  Release: %s, Warning: Error deleting %s hook %s based on delete policy, continue", name, hook, h.Hook.Name)
			}
		}
	}()

	return createdHooks, executingHooks, nil
}

func (hc *HookClient) deleteHookByPolicy(h *helm.Hook, policy release.HookDeletePolicy, name string, hook release.HookEvent, k8sClient KubernetesClient) error {
	rss := helm.KubernetesResource{
		GVK:  h.KRT.GVK,
		Name: h.Hook.Name,
	}
	if hookHasDeletePolicy(h, policy) {
		log.Printf("  Release: %s, Deleting %s hook %s due to %q policy", name, hook, h.Hook.Name, policy)
		if errHookDelete := k8sClient.deleteResources(append([]helm.KubernetesResource{}, rss), hc.kubeNameSpace); errHookDelete != nil {
			if strings.Contains(errHookDelete.Error(), "not found") {
				return nil
			} else {
				log.Printf("  Release: %s, Warning: %s hook %s, filePath %s could not be deleted: %s", name, hook, h.Hook.Name, h.KRT.FilePath ,errHookDelete)
				return errHookDelete
			}
		} else {
			//Verify that the rss is deleted
			isDeleted := false
			for !isDeleted {
				log.Printf("  Release: %s, Waiting on deleting %s hook %s for release %s due to %q policy", name, hook, h.Hook.Name, name, policy)
				if _, err := k8sClient.GetResourceStatus(rss, hc.kubeNameSpace); err != nil {
					if strings.Contains(err.Error(), "not found") {
						log.Printf("  Release: %s, Deleted %s hook %s for release %s due to %q policy", name, hook, h.Hook.Name, name, policy)
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