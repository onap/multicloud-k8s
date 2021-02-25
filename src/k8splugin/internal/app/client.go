/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Samsung Electronics

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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesClient encapsulates the different clients' interfaces
// we need when interacting with a Kubernetes cluster
type KubernetesClient struct {
	rawConfig      clientcmd.ClientConfig
	restConfig     *rest.Config
	clientSet      kubernetes.Interface
	dynamicClient  dynamic.Interface
	discoverClient *disk.CachedDiscoveryClient
	restMapper     meta.RESTMapper
	instanceID     string
}

// ResourceStatus holds Resource Runtime Data
type ResourceStatus struct {
	Name   string                    `json:"name"`
	GVK    schema.GroupVersionKind   `json:"GVK"`
	Status unstructured.Unstructured `json:"status"`
}

// getPodsByLabel yields status of all pods under given instance ID
func (k *KubernetesClient) getPodsByLabel(namespace string) ([]ResourceStatus, error) {
	client := k.GetStandardClient().CoreV1().Pods(namespace)
	listOpts := metav1.ListOptions{
		LabelSelector: config.GetConfiguration().KubernetesLabelName + "=" + k.instanceID,
	}
	podList, err := client.List(listOpts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Retrieving PodList from cluster")
	}
	resp := make([]ResourceStatus, 0, len(podList.Items))
	cumulatedErrorMsg := make([]string, 0)
	for _, pod := range podList.Items {
		podContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
		if err != nil {
			cumulatedErrorMsg = append(cumulatedErrorMsg, err.Error())
			continue
		}
		var unstrPod unstructured.Unstructured
		unstrPod.SetUnstructuredContent(podContent)
		podStatus := ResourceStatus{
			Name:   unstrPod.GetName(),
			GVK:    schema.FromAPIVersionAndKind("v1", "Pod"),
			Status: unstrPod,
		}
		resp = append(resp, podStatus)
	}
	if len(cumulatedErrorMsg) != 0 {
		return resp, pkgerrors.New("Converting podContent to unstruct error:\n" +
			strings.Join(cumulatedErrorMsg, "\n"))
	}
	return resp, nil
}

func (k *KubernetesClient) queryResources(apiVersion, kind, labelSelector, namespace string) ([]ResourceStatus, error) {
	dynClient := k.GetDynamicClient()
	mapper := k.GetMapper()
	gvk := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Preparing mapper based on GVK")
	}

	gvr := mapping.Resource
	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	var unstrList *unstructured.UnstructuredList
	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		unstrList, err = dynClient.Resource(gvr).Namespace(namespace).List(opts)
	case meta.RESTScopeNameRoot:
		unstrList, err = dynClient.Resource(gvr).List(opts)
	default:
		return nil, pkgerrors.New("Got an unknown RESTScopeName for mapping: " + gvk.String())
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Querying for resources")
	}

	resp := make([]ResourceStatus, len(unstrList.Items))
	for _, unstr := range unstrList.Items {
		resp = append(resp, ResourceStatus{unstr.GetName(), gvk, unstr})
	}
	return resp, nil
}

// getResourcesStatus yields status of given generic resource
func (k *KubernetesClient) getResourceStatus(res helm.KubernetesResource, namespace string) (ResourceStatus, error) {
	dynClient := k.GetDynamicClient()
	mapper := k.GetMapper()
	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: res.GVK.Group,
		Kind:  res.GVK.Kind,
	}, res.GVK.Version)
	if err != nil {
		return ResourceStatus{},
			pkgerrors.Wrap(err, "Preparing mapper based on GVK")
	}

	gvr := mapping.Resource
	opts := metav1.GetOptions{}
	var unstruct *unstructured.Unstructured
	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		unstruct, err = dynClient.Resource(gvr).Namespace(namespace).Get(res.Name, opts)
	case meta.RESTScopeNameRoot:
		unstruct, err = dynClient.Resource(gvr).Get(res.Name, opts)
	default:
		return ResourceStatus{}, pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + res.GVK.String())
	}

	if err != nil {
		return ResourceStatus{}, pkgerrors.Wrap(err, "Getting object status")
	}

	return ResourceStatus{unstruct.GetName(), res.GVK, *unstruct}, nil
}

// getKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the cloudregion. This is written out to a file.
func (k *KubernetesClient) getKubeConfig(cloudregion string) (string, error) {

	conn := connection.NewConnectionClient()
	kubeConfigPath, err := conn.Download(cloudregion)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Downloading kubeconfig")
	}

	return kubeConfigPath, nil
}

// Init loads the Kubernetes configuation values stored into the local configuration file
func (k *KubernetesClient) Init(cloudregion string, iid string) error {
	if cloudregion == "" {
		return pkgerrors.New("Cloudregion is empty")
	}

	if iid == "" {
		return pkgerrors.New("Instance ID is empty")
	}

	k.instanceID = iid

	configPath, err := k.getKubeConfig(cloudregion)
	if err != nil {
		return pkgerrors.Wrap(err, "Get kubeconfig file")
	}

	//Remove kubeconfigfile after the clients are created
	defer os.Remove(configPath)

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}

	k.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	k.dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating dynamic client")
	}

	k.discoverClient, err = disk.NewCachedDiscoveryClientForConfig(config, os.TempDir(), "", 10*time.Minute)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating discovery client")
	}

	k.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(k.discoverClient)
	k.restConfig = config

	//Spawn ClientConfig
	kubeFile, err := os.Open(configPath)
	if err != nil {
		return pkgerrors.Wrap(err, "Opening kubeConfig")
	}
	kubeData, err := ioutil.ReadAll(kubeFile)
	if err != nil {
		return pkgerrors.Wrap(err, "Reading kubeConfig")
	}
	k.rawConfig, err = clientcmd.NewClientConfigFromBytes(kubeData)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating rawConfig")
	}

	return nil
}

func (k *KubernetesClient) ensureNamespace(namespace string) error {

	pluginImpl, err := plugin.GetPluginByKind("Namespace")
	if err != nil {
		return pkgerrors.Wrap(err, "Loading Namespace Plugin")
	}

	ns, err := pluginImpl.Get(helm.KubernetesResource{
		Name: namespace,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Namespace",
		},
	}, namespace, k)

	// Check for errors getting the namespace while ignoring errors where the namespace does not exist
	// Error message when namespace does not exist: "namespaces "namespace-name" not found"
	if err != nil && strings.Contains(err.Error(), "not found") == false {
		log.Error("Error checking for namespace", log.Fields{
			"error":     err,
			"namespace": namespace,
		})
		return pkgerrors.Wrap(err, "Error checking for namespace: "+namespace)
	}

	if ns == "" {
		log.Info("Creating Namespace", log.Fields{
			"namespace": namespace,
		})

		_, err = pluginImpl.Create("", namespace, k)
		if err != nil {
			log.Error("Error Creating Namespace", log.Fields{
				"error":     err,
				"namespace": namespace,
			})
			return pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}
	return nil
}

func (k *KubernetesClient) createKind(resTempl helm.KubernetesResourceTemplate,
	namespace string) (helm.KubernetesResource, error) {

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + "does not exists")
	}

	log.Info("Processing Kubernetes Resource", log.Fields{
		"filepath": resTempl.FilePath,
	})

	pluginImpl, err := plugin.GetPluginByKind(resTempl.GVK.Kind)
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error loading plugin")
	}

	createdResourceName, err := pluginImpl.Create(resTempl.FilePath, namespace, k)
	if err != nil {
		log.Error("Error Creating Resource", log.Fields{
			"error":    err,
			"gvk":      resTempl.GVK,
			"filepath": resTempl.FilePath,
		})
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error in plugin "+resTempl.GVK.Kind+" plugin")
	}

	log.Info("Created Kubernetes Resource", log.Fields{
		"resource": createdResourceName,
		"gvk":      resTempl.GVK,
	})

	return helm.KubernetesResource{
		GVK:  resTempl.GVK,
		Name: createdResourceName,
	}, nil
}

func (k *KubernetesClient) updateKind(resTempl helm.KubernetesResourceTemplate,
	namespace string) (helm.KubernetesResource, error) {

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + "does not exists")
	}

	log.Info("Processing Kubernetes Resource", log.Fields{
		"filepath": resTempl.FilePath,
	})

	pluginImpl, err := plugin.GetPluginByKind(resTempl.GVK.Kind)
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error loading plugin")
	}

	updatedResourceName, err := pluginImpl.Update(resTempl.FilePath, namespace, k)
	if err != nil {
		log.Error("Error Updating Resource", log.Fields{
			"error":    err,
			"gvk":      resTempl.GVK,
			"filepath": resTempl.FilePath,
		})
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error in plugin "+resTempl.GVK.Kind+" plugin")
	}

	log.Info("Updated Kubernetes Resource", log.Fields{
		"resource": updatedResourceName,
		"gvk":      resTempl.GVK,
	})

	return helm.KubernetesResource{
		GVK:  resTempl.GVK,
		Name: updatedResourceName,
	}, nil
}

func (k *KubernetesClient) createResources(sortedTemplates []helm.KubernetesResourceTemplate,
	namespace string) ([]helm.KubernetesResource, error) {

	err := k.ensureNamespace(namespace)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Creating Namespace")
	}

	var createdResources []helm.KubernetesResource
	for _, resTempl := range sortedTemplates {
		resCreated, err := k.createKind(resTempl, namespace)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "Error creating kind: %+v", resTempl.GVK)
		}
		createdResources = append(createdResources, resCreated)
	}

	return createdResources, nil
}

func (k *KubernetesClient) updateResources(sortedTemplates []helm.KubernetesResourceTemplate,
	namespace string) ([]helm.KubernetesResource, error) {

	err := k.ensureNamespace(namespace)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Creating Namespace")
	}

	var updatedResources []helm.KubernetesResource
	for _, resTempl := range sortedTemplates {
		resUpdated, err := k.updateKind(resTempl, namespace)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "Error updating kind: %+v", resTempl.GVK)
		}
		updatedResources = append(updatedResources, resUpdated)
	}

	return updatedResources, nil
}

func (k *KubernetesClient) deleteKind(resource helm.KubernetesResource, namespace string) error {
	log.Warn("Deleting Resource", log.Fields{
		"gvk":      resource.GVK,
		"resource": resource.Name,
	})

	pluginImpl, err := plugin.GetPluginByKind(resource.GVK.Kind)
	if err != nil {
		return pkgerrors.Wrap(err, "Error loading plugin")
	}

	err = pluginImpl.Delete(resource, namespace, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting "+resource.Name)
	}

	return nil
}

func (k *KubernetesClient) deleteResources(resources []helm.KubernetesResource, namespace string) error {
	//TODO: Investigate if deletion should be in a particular order
	for _, res := range resources {
		err := k.deleteKind(res, namespace)
		if err != nil {
			return pkgerrors.Wrap(err, "Deleting resources")
		}
	}

	return nil
}

//GetMapper returns the RESTMapper that was created for this client
func (k *KubernetesClient) GetMapper() meta.RESTMapper {
	return k.restMapper
}

//GetDynamicClient returns the dynamic client that is needed for
//unstructured REST calls to the apiserver
func (k *KubernetesClient) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}

// GetStandardClient returns the standard client that can be used to handle
// standard kubernetes kinds
func (k *KubernetesClient) GetStandardClient() kubernetes.Interface {
	return k.clientSet
}

//GetInstanceID returns the instanceID that is injected into all the
//resources created by the plugin
func (k *KubernetesClient) GetInstanceID() string {
	return k.instanceID
}

//Following set of methods are implemented so that KubernetesClient
//implements genericclioptions.RESTClientGetter interface
func (k *KubernetesClient) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return k.discoverClient, nil
}
func (k *KubernetesClient) ToRESTMapper() (meta.RESTMapper, error) {
	return k.GetMapper(), nil
}
func (k *KubernetesClient) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return k.rawConfig
}
func (k *KubernetesClient) ToRESTConfig() (*rest.Config, error) {
	return k.restConfig, nil
}
