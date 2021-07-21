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
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/kube"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	//appsv1beta1 "k8s.io/api/apps/v1beta1"
	//appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	//extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	//apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	//apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"os"
	"strings"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
	logger "log"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	cachetools "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
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
	readyChecker   kube.ReadyChecker
}

// ResourceStatus holds Resource Runtime Data
type ResourceStatus struct {
	Name   string                    `json:"name"`
	GVK    schema.GroupVersionKind   `json:"GVK"`
	Status unstructured.Unstructured `json:"status"`
}

func (k *KubernetesClient) getObjType(kind string) (runtime.Object, error) {
	switch kind {
	case "Job":
		return &batchv1.Job{}, nil
	case "Pod":
		return &corev1.Pod{}, nil
	case "Deployment":
		return &appsv1.Deployment{}, nil
	case "DaemonSet":
		return &appsv1.DaemonSet{}, nil
	case "StatefulSet":
		return &appsv1.StatefulSet{}, nil
	default:
		return nil, pkgerrors.New("kind " + kind + " unknown")
	}
}

func (k *KubernetesClient) getRestApi(apiVersion string) (rest.Interface, error) {
	//based on kubectl api-versions
	switch apiVersion {
	case "admissionregistration.k8s.io/v1":
		return k.clientSet.AdmissionregistrationV1().RESTClient(), nil
	case "admissionregistration.k8s.io/v1beta1":
		return k.clientSet.AdmissionregistrationV1beta1().RESTClient(), nil
	case "apps/v1":
		return k.clientSet.AppsV1().RESTClient(), nil
	case "apps/v1beta1":
		return k.clientSet.AppsV1beta1().RESTClient(), nil
	case "apps/v1beta2":
		return k.clientSet.AppsV1beta2().RESTClient(), nil
	case "authentication.k8s.io/v1":
		return k.clientSet.AuthenticationV1().RESTClient(), nil
	case "authentication.k8s.io/v1beta1":
		return k.clientSet.AuthenticationV1beta1().RESTClient(), nil
	case "authorization.k8s.io/v1":
		return k.clientSet.AuthorizationV1().RESTClient(), nil
	case "authorization.k8s.io/v1beta1":
		return k.clientSet.AuthorizationV1beta1().RESTClient(), nil
	case "autoscaling/v1":
		return k.clientSet.AutoscalingV1().RESTClient(), nil
	case "autoscaling/v2beta1":
		return k.clientSet.AutoscalingV2beta1().RESTClient(), nil
	case "autoscaling/v2beta2":
		return k.clientSet.AutoscalingV2beta2().RESTClient(), nil
	case "batch/v1":
		return k.clientSet.BatchV1().RESTClient(), nil
	case "batch/v1beta1":
		return k.clientSet.BatchV1beta1().RESTClient(), nil
	case "certificates.k8s.io/v1":
		return k.clientSet.CertificatesV1().RESTClient(), nil
	case "certificates.k8s.io/v1beta1":
		return k.clientSet.CertificatesV1beta1().RESTClient(), nil
	case "coordination.k8s.io/v1":
		return k.clientSet.CoordinationV1().RESTClient(), nil
	case "coordination.k8s.io/v1beta1":
		return k.clientSet.CoordinationV1beta1().RESTClient(), nil
	case "v1":
		return k.clientSet.CoreV1().RESTClient(), nil
	case "discovery.k8s.io/v1":
		return k.clientSet.DiscoveryV1().RESTClient(), nil
	case "discovery.k8s.io/v1beta1":
		return k.clientSet.DiscoveryV1beta1().RESTClient(), nil
	case "events.k8s.io/v1":
		return k.clientSet.EventsV1().RESTClient(), nil
	case "events.k8s.io/v1beta1":
		return k.clientSet.EventsV1beta1().RESTClient(), nil
	case "extensions/v1beta1":
		return k.clientSet.ExtensionsV1beta1().RESTClient(), nil
	case "flowcontrol.apiserver.k8s.io/v1alpha1":
		return k.clientSet.FlowcontrolV1alpha1().RESTClient(), nil
	case "networking.k8s.io/v1":
		return k.clientSet.NetworkingV1().RESTClient(), nil
	case "networking.k8s.io/v1beta1":
		return k.clientSet.NetworkingV1beta1().RESTClient(), nil
	case "node.k8s.io/v1alpha1":
		return k.clientSet.NodeV1alpha1().RESTClient(), nil
	case "node.k8s.io/v1beta1":
		return k.clientSet.NodeV1beta1().RESTClient(), nil
	case "policy/v1beta1":
		return k.clientSet.PolicyV1beta1().RESTClient(), nil
	case "rbac.authorization.k8s.io/v1":
		return k.clientSet.RbacV1().RESTClient(), nil
	case "rbac.authorization.k8s.io/v1alpha1":
		return k.clientSet.RbacV1alpha1().RESTClient(), nil
	case "rbac.authorization.k8s.io/v1beta1":
		return k.clientSet.RbacV1beta1().RESTClient(), nil
	case "scheduling.k8s.io/v1":
		return k.clientSet.SchedulingV1().RESTClient(), nil
	case "scheduling.k8s.io/v1alpha1":
		return k.clientSet.SchedulingV1alpha1().RESTClient(), nil
	case "scheduling.k8s.io/v1beta1":
		return k.clientSet.SchedulingV1beta1().RESTClient(), nil
	case "storage.k8s.io/v1":
		return k.clientSet.StorageV1().RESTClient(), nil
	case "storage.k8s.io/v1alpha1":
		return k.clientSet.StorageV1alpha1().RESTClient(), nil
	case "storage.k8s.io/v1beta1":
		return k.clientSet.StorageV1beta1().RESTClient(), nil
	default:
		return nil, pkgerrors.New("Api version " + apiVersion + " unknown")
	}
}

func (k *KubernetesClient) watchUntilReady(timeout time.Duration, ns string, res helm.KubernetesResource) error {
	switch res.GVK.Kind {
	case "Job", "Pod", "Deployment", "DaemonSet", "StatefulSet":
	default:
		return nil
	}
	selector, err := fields.ParseSelector(fmt.Sprintf("metadata.name=%s", res.Name))
	if err != nil {
		return err
	}

	mapper := k.GetMapper()
	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: res.GVK.Group,
		Kind:  res.GVK.Kind,
	}, res.GVK.Version)
	if err != nil {
		return pkgerrors.Wrap(err, "Preparing mapper based on GVK")
	}
	apiVersion, kind := res.GVK.ToAPIVersionAndKind()
	logger.Printf("apiVersion: %s, Kind: %s", apiVersion, kind)
	restClient, err := k.getRestApi(apiVersion)
	if err != nil {
		return pkgerrors.Wrap(err, "Get rest client")
	}
	lw := cachetools.NewListWatchFromClient(restClient, mapping.Resource.Resource, ns, selector)

	// What we watch for depends on the Kind.
	// - For a Job, we watch for completion.
	// - For all else, we watch until Ready.
	// In the future, we might want to add some special logic for types
	// like Ingress, Volume, etc.
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()
	objType, err := k.getObjType(kind)
	if err != nil {
		return pkgerrors.Wrap(err, "Get obj type")
	}
	_, err = watchtools.UntilWithSync(ctx, lw, objType, nil, func(e watch.Event) (bool, error) {
		obj := e.Object
		switch e.Type {
		case watch.Added, watch.Modified:
			// For things like a secret or a config map, this is the best indicator
			// we get. We care mostly about jobs, where what we want to see is
			// the status go into a good state.
			logger.Printf("Add/Modify event for %s: %v", res.Name, e.Type)
			switch res.GVK.Kind {
			case "Job":
				return k.waitForJob(obj, res.Name)
			case "Pod":
				return k.waitForPodSuccess(obj, res.Name)
			case "Deployment":
				return k.waitForDeploymentSuccess(obj, res.Name)
			case "DaemonSet":
				return k.waitForDaemonSetSuccess(obj, res.Name)
			case "StatefulSet":
				return k.waitForStatefulSetSuccess(obj, res.Name)
			}
			return true, nil
		case watch.Deleted:
			logger.Printf("Deleted event for %s", res.Name)
			return true, nil
		case watch.Error:
			// Handle error and return with an error.
			logger.Printf("Error event for %s", res.Name)
			return true, pkgerrors.New("failed to deploy " + res.Name)
		default:
			return false, nil
		}
	})
	logger.Printf("Done for %s", res.Name)
	return nil
}

// waitForJob is a helper that waits for a job to complete.
//
// This operates on an event returned from a watcher.
func (k *KubernetesClient) waitForJob(obj runtime.Object, name string) (bool, error) {
	o, ok := obj.(*batchv1.Job)
	if !ok {
		return true, pkgerrors.New("expected " + name + " to be a *batch.Job, got " + obj.GetObjectKind().GroupVersionKind().Kind)
	}

	for _, c := range o.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == "True" {
			return true, nil
		} else if c.Type == batchv1.JobFailed && c.Status == "True" {
			return true, pkgerrors.New("job failed: " + c.Reason)
		}
	}

	logger.Printf("%s: Jobs active: %d, jobs failed: %d, jobs succeeded: %d", name, o.Status.Active, o.Status.Failed, o.Status.Succeeded)
	return false, nil
}

// waitForPodSuccess is a helper that waits for a pod to complete.
//
// This operates on an event returned from a watcher.
func (k *KubernetesClient) waitForPodSuccess(obj runtime.Object, name string) (bool, error) {
	o, ok := obj.(*corev1.Pod)
	if !ok {
		return true, pkgerrors.New("expected " + name + " to be a *v1.Pod, got " + obj.GetObjectKind().GroupVersionKind().Kind)
	}

	switch o.Status.Phase {
	case corev1.PodSucceeded:
		logger.Printf("Pod %s succeeded", o.Name)
		return true, nil
	case corev1.PodFailed:
		return true, pkgerrors.New("pod " + o.Name + " failed")
	case corev1.PodPending:
		logger.Printf("Pod %s pending", o.Name)
	case corev1.PodRunning:
		logger.Printf("Pod %s running", o.Name)
	}

	return false, nil
}

// waitForDeploymentSuccess is a helper that waits for a deployment to run.
//
// This operates on an event returned from a watcher.
func (k *KubernetesClient) waitForDeploymentSuccess(obj runtime.Object, name string) (bool, error) {
	o, ok := obj.(*appsv1.Deployment)
	if !ok {
		return true, pkgerrors.New("expected " + name + " to be a *apps.Deployment, got " + obj.GetObjectKind().GroupVersionKind().Kind)
	}

	// If paused deployment will never be ready -> consider ready
	if o.Spec.Paused {
		logger.Printf("Depoyment %s is paused, consider ready", o.Name)
		return true, nil
	}

	// Find RS associated with deployment
	newReplicaSet, err := GetNewReplicaSet(o, k.clientSet.AppsV1())
	if err != nil || newReplicaSet == nil {
		return false, err
	}
	expectedReady := *o.Spec.Replicas - MaxUnavailable(*o)
	if !(newReplicaSet.Status.ReadyReplicas >= expectedReady) {
		logger.Printf("Deployment is not ready: %s/%s. %d out of %d expected pods are ready", o.Namespace, o.Name, newReplicaSet.Status.ReadyReplicas, expectedReady)
		return false, nil
	}
	return true, nil
}

// waitForDaemonSetSuccess is a helper that waits for a daemonSet to run.
//
// This operates on an event returned from a watcher.
func (k *KubernetesClient) waitForDaemonSetSuccess(obj runtime.Object, name string) (bool, error) {
	o, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return true, pkgerrors.New("expected " + name + " to be a *apps.DaemonSet, got " + obj.GetObjectKind().GroupVersionKind().Kind)
	}

	// If the update strategy is not a rolling update, there will be nothing to wait for
	if o.Spec.UpdateStrategy.Type != appsv1.RollingUpdateDaemonSetStrategyType {
		return true, nil
	}

	// Make sure all the updated pods have been scheduled
	if o.Status.UpdatedNumberScheduled != o.Status.DesiredNumberScheduled {
		logger.Printf("DaemonSet is not ready: %s/%s. %d out of %d expected pods have been scheduled", o.Namespace, o.Name, o.Status.UpdatedNumberScheduled, o.Status.DesiredNumberScheduled)
		return false, nil
	}
	maxUnavailable, err := intstr.GetValueFromIntOrPercent(o.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable, int(o.Status.DesiredNumberScheduled), true)
	if err != nil {
		// If for some reason the value is invalid, set max unavailable to the
		// number of desired replicas. This is the same behavior as the
		// `MaxUnavailable` function in deploymentutil
		maxUnavailable = int(o.Status.DesiredNumberScheduled)
	}

	expectedReady := int(o.Status.DesiredNumberScheduled) - maxUnavailable
	if !(int(o.Status.NumberReady) >= expectedReady) {
		logger.Printf("DaemonSet is not ready: %s/%s. %d out of %d expected pods are ready", o.Namespace, o.Name, o.Status.NumberReady, expectedReady)
		return false, nil
	}
	return true, nil
}

// waitForStatefulSetSuccess is a helper that waits for a statefulSet to run.
//
// This operates on an event returned from a watcher.
func (k *KubernetesClient) waitForStatefulSetSuccess(obj runtime.Object, name string) (bool, error) {
	o, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return true, pkgerrors.New("expected " + name + " to be a *apps.StatefulSet, got " + obj.GetObjectKind().GroupVersionKind().Kind)
	}

	// If the update strategy is not a rolling update, there will be nothing to wait for
	if o.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		return true, nil
	}

	// Dereference all the pointers because StatefulSets like them
	var partition int
	// 1 is the default for replicas if not set
	var replicas = 1
	// For some reason, even if the update strategy is a rolling update, the
	// actual rollingUpdate field can be nil. If it is, we can safely assume
	// there is no partition value
	if o.Spec.UpdateStrategy.RollingUpdate != nil && o.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		partition = int(*o.Spec.UpdateStrategy.RollingUpdate.Partition)
	}
	if o.Spec.Replicas != nil {
		replicas = int(*o.Spec.Replicas)
	}

	// Because an update strategy can use partitioning, we need to calculate the
	// number of updated replicas we should have. For example, if the replicas
	// is set to 3 and the partition is 2, we'd expect only one pod to be
	// updated
	expectedReplicas := replicas - partition

	// Make sure all the updated pods have been scheduled
	if int(o.Status.UpdatedReplicas) != expectedReplicas {
		logger.Printf("StatefulSet is not ready: %s/%s. %d out of %d expected pods have been scheduled", o.Namespace, o.Name, o.Status.UpdatedReplicas, expectedReplicas)
		return false, nil
	}

	if int(o.Status.ReadyReplicas) != replicas {
		logger.Printf("StatefulSet is not ready: %s/%s. %d out of %d expected pods are ready", o.Namespace, o.Name, o.Status.ReadyReplicas, replicas)
		return false, nil
	}
	return true, nil
}

// getPodsByLabel yields status of all pods under given instance ID
func (k *KubernetesClient) getPodsByLabel(namespace string) ([]ResourceStatus, error) {
	client := k.GetStandardClient().CoreV1().Pods(namespace)
	listOpts := metav1.ListOptions{
		LabelSelector: config.GetConfiguration().KubernetesLabelName + "=" + k.instanceID,
	}
	podList, err := client.List(context.TODO(), listOpts)
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
		unstrList, err = dynClient.Resource(gvr).Namespace(namespace).List(context.TODO(), opts)
	case meta.RESTScopeNameRoot:
		unstrList, err = dynClient.Resource(gvr).List(context.TODO(), opts)
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

// GetResourcesStatus yields status of given generic resource
func (k *KubernetesClient) GetResourceStatus(res helm.KubernetesResource, namespace string) (ResourceStatus, error) {
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
		unstruct, err = dynClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), res.Name, opts)
	case meta.RESTScopeNameRoot:
		unstruct, err = dynClient.Resource(gvr).Get(context.TODO(), res.Name, opts)
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

	k.readyChecker = kube.NewReadyChecker(k.clientSet, logger.Printf, kube.PausedAsReady(true), kube.CheckJobs(true))

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

func (k *KubernetesClient) CreateKind(resTempl helm.KubernetesResourceTemplate, namespace string) (helm.KubernetesResource, error) {

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + " does not exists")
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
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + " does not exists")
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
		resCreated, err := k.CreateKind(resTempl, namespace)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "Error creating kind: %+v", resTempl.GVK)
		}
		createdResources = append(createdResources, resCreated)
	}

	return createdResources, nil
}

func (k *KubernetesClient) CreateHookResources(h *helm.Hook, namespace string) (*helm.KubernetesResource, error) {
	err := k.ensureNamespace(namespace)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Creating Namespace")
	}

	resTempl := helm.KubernetesResourceTemplate{
		GVK:      h.KRT.GVK,
		FilePath: h.KRT.FilePath,
	}
	resCreated, err := k.CreateKind(resTempl, namespace)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "Error creating kind: %+v", resTempl.GVK)
	}
	return &resCreated, nil
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

func (k *KubernetesClient) DeleteKind(resource helm.KubernetesResource, namespace string) error {
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
		err := k.DeleteKind(res, namespace)
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
