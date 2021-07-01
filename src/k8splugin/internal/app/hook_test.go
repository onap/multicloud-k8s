/*
Copyright © 2021 Nokia Bell Labs.
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
	"encoding/base64"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func generateHookList() []*helm.Hook {
	var hookList []*helm.Hook
	preInstallHook1 := helm.Hook{
		Hook: release.Hook{
			Name : "preinstall1",
			Kind : "Job",
			Path : "",
			Manifest : "",
			Events : []release.HookEvent{release.HookPreInstall},
			LastRun : release.HookExecution{
				StartedAt:   time.Now(),
				CompletedAt: time.Now(),
				Phase:       "",
			},
			Weight : -5,
			DeletePolicies : []release.HookDeletePolicy{},
		},
		KRT:  helm.KubernetesResourceTemplate{
			GVK: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			FilePath: "../../mock_files/mock_yamls/job.yaml",
		},
	}
	preInstallHook2 := helm.Hook{
		Hook: release.Hook{
			Name : "preinstall2",
			Kind : "Deployment",
			Path : "",
			Manifest : "",
			Events : []release.HookEvent{release.HookPreInstall},
			LastRun : release.HookExecution{
				StartedAt:   time.Now(),
				CompletedAt: time.Now(),
				Phase:       "",
			},
			Weight : 0,
			DeletePolicies : []release.HookDeletePolicy{},
		},
		KRT:  helm.KubernetesResourceTemplate{
			GVK: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			FilePath: "../../mock_files/mock_yamls/job.yaml",
		},
	}
	postInstallHook := helm.Hook{
		Hook: release.Hook{
			Name : "postinstall",
			Kind : "Job",
			Path : "",
			Manifest : "",
			Events : []release.HookEvent{release.HookPostInstall},
			LastRun : release.HookExecution{
				StartedAt:   time.Now(),
				CompletedAt: time.Now(),
				Phase:       "",
			},
			Weight : -5,
			DeletePolicies : []release.HookDeletePolicy{},
		},
		KRT:  helm.KubernetesResourceTemplate{
			GVK: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			FilePath: "../../mock_files/mock_yamls/job.yaml",
		},
	}
	preDeleteHook := helm.Hook{
		Hook: release.Hook{
			Name : "predelete",
			Kind : "Job",
			Path : "",
			Manifest : "",
			Events : []release.HookEvent{release.HookPreDelete},
			LastRun : release.HookExecution{
				StartedAt:   time.Now(),
				CompletedAt: time.Now(),
				Phase:       "",
			},
			Weight : -5,
			DeletePolicies : []release.HookDeletePolicy{},
		},
		KRT:  helm.KubernetesResourceTemplate{
			GVK: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			FilePath: "../../mock_files/mock_yamls/job.yaml",
		},
	}
	postDeleteHook := helm.Hook{
		Hook: release.Hook{
			Name : "postdelete",
			Kind : "Job",
			Path : "",
			Manifest : "",
			Events : []release.HookEvent{release.HookPostDelete},
			LastRun : release.HookExecution{
				StartedAt:   time.Now(),
				CompletedAt: time.Now(),
				Phase:       "",
			},
			Weight : -5,
			DeletePolicies : []release.HookDeletePolicy{},
		},
		KRT:  helm.KubernetesResourceTemplate{
			GVK: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			FilePath: "../../mock_files/mock_yamls/job.yaml",
		},
	}
	hookList = append(hookList, &preInstallHook2)
	hookList = append(hookList, &preInstallHook1)
	hookList = append(hookList, &postInstallHook)
	hookList = append(hookList, &preDeleteHook)
	hookList = append(hookList, &postDeleteHook)

	return hookList
}

func TestGetHookByEvent(t *testing.T) {
	hookList := generateHookList()
	hookClient := NewHookClient("test", "test", "rbdef", "instance")
	t.Run("Get pre-install hook", func(t *testing.T) {
		preinstallList := hookClient.getHookByEvent(hookList, release.HookPreInstall)
		if len(preinstallList) != 2 {
			t.Fatalf("TestGetHookByEvent error: expected=2 preinstall hook, result= %d", len(preinstallList))
		}
		if preinstallList[0].Hook.Name != "preinstall2" {
			t.Fatalf("TestGetHookByEvent error: expect name of 1st preinstall hook is preinstall2, result= %s", preinstallList[0].Hook.Name)
		}
		if preinstallList[1].Hook.Name != "preinstall1" {
			t.Fatalf("TestGetHookByEvent error: expect name of 2nd preinstall hook is preinstall1, result= %s", preinstallList[0].Hook.Name)
		}
	})
	t.Run("Get post-install hook", func(t *testing.T) {
		postinstallList := hookClient.getHookByEvent(hookList, release.HookPostInstall)
		if len(postinstallList) != 1 {
			t.Fatalf("TestGetHookByEvent error: expected=1 postinstall hook, result= %d", len(postinstallList))
		}
		if postinstallList[0].Hook.Name != "postinstall" {
			t.Fatalf("TestGetHookByEvent error: expect name of 1st postinstall hook is postinstall, result= %s", postinstallList[0].Hook.Name)
		}
	})
	t.Run("Get pre-delete hook", func(t *testing.T) {
		predeleteList := hookClient.getHookByEvent(hookList, release.HookPreDelete)
		if len(predeleteList) != 1 {
			t.Fatalf("TestGetHookByEvent error: expected=1 predelete hook, result= %d", len(predeleteList))
		}
		if predeleteList[0].Hook.Name != "predelete" {
			t.Fatalf("TestGetHookByEvent error: expect name of 1st predelete hook is predelete, result= %s", predeleteList[0].Hook.Name)
		}
	})
	t.Run("Get post-delete hook", func(t *testing.T) {
		postdeleteList := hookClient.getHookByEvent(hookList, release.HookPostDelete)
		if len(postdeleteList) != 1 {
			t.Fatalf("TestGetHookByEvent error: expected=1 postdelete hook, result= %d", len(postdeleteList))
		}
		if postdeleteList[0].Hook.Name != "postdelete" {
			t.Fatalf("TestGetHookByEvent error: expect name of 1st postdelete hook is postdelete, result= %s", postdeleteList[0].Hook.Name)
		}
	})
}

func TestShortHook(t *testing.T) {
	hookList := generateHookList()
	hookClient := NewHookClient("test", "test", "rbdef", "instance")
	preinstallList := hookClient.getHookByEvent(hookList, release.HookPreInstall)
	t.Run("Short pre-install hook", func(t *testing.T) {
		shortedHooks := sortByHookWeight(preinstallList)
		if shortedHooks[0].Hook.Name != "preinstall1" {
			t.Fatalf("TestShortHook error: expect name of 1st preinstall hook is preinstall1, result= %s", preinstallList[0].Hook.Name)
		}
		if shortedHooks[1].Hook.Name != "preinstall2" {
			t.Fatalf("TestShortHook error: expect name of 2nd preinstall hook is preinstall2, result= %s", preinstallList[0].Hook.Name)
		}
	})
}

func TestExecHook(t *testing.T) {
	hookList := generateHookList()
	hookClient := NewHookClient("test", "test", "rbdef", "instance")
	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	// Load the mock kube config file into memory
	fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
	if err != nil {
		t.Fatal("Unable to read mock_kube_config")
	}
	db.DBconn = &db.MockDB{
		Items: map[string]map[string][]byte{
			connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
				"metadata": []byte(
					"{\"cloud-region\":\"mock_connection\"," +
						"\"cloud-owner\":\"mock_owner\"," +
						"\"kubeconfig\": \"" + base64.StdEncoding.EncodeToString(fd) + "\"}"),
			},
		},
	}

	k8sClient := KubernetesClient{}
	err = k8sClient.Init("mock_connection", "test")
	if err != nil {
		t.Fatal(err.Error())
	}
	err = hookClient.ExecHook(k8sClient, hookList, release.HookPreInstall,10,0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = hookClient.ExecHook(k8sClient, hookList, release.HookPostInstall,10,0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = hookClient.ExecHook(k8sClient, hookList, release.HookPreDelete,10,0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = hookClient.ExecHook(k8sClient, hookList, release.HookPostDelete,10,0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
}