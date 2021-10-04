/*
Copyright 2018 Intel Corporation.
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

package main

import (
	"context"
	"fmt"
	logger "log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	//appsv1beta1 "k8s.io/api/apps/v1beta1"
	//appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	pkgerrors "github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"
	cachetools "k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

// Compile time check to see if genericPlugin implements the correct interface
var _ plugin.Reference = genericPlugin{}

// ExportedVariable is what we will look for when calling the generic plugin
var ExportedVariable genericPlugin

type genericPlugin struct {
}

func (g genericPlugin) WatchUntilReady(
	timeout time.Duration,
	ns string,
	res helm.KubernetesResource,
	mapper meta.RESTMapper,
	restClient rest.Interface,
	objType runtime.Object,
	clientSet kubernetes.Interface) error {
	selector, err := fields.ParseSelector(fmt.Sprintf("metadata.name=%s", res.Name))
	if err != nil {
		return err
	}

	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: res.GVK.Group,
		Kind:  res.GVK.Kind,
	}, res.GVK.Version)
	if err != nil {
		return pkgerrors.Wrap(err, "Preparing mapper based on GVK")
	}
	lw := cachetools.NewListWatchFromClient(restClient, mapping.Resource.Resource, ns, selector)

	// What we watch for depends on the Kind.
	// - For a Job, we watch for completion.
	// - For all else, we watch until Ready.
	// In the future, we might want to add some special logic for types
	// like Ingress, Volume, etc.
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

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
				return g.waitForJob(obj, res.Name)
			case "Pod":
				return g.waitForPodSuccess(obj, res.Name)
			case "Deployment":
				return g.waitForDeploymentSuccess(obj, res.Name, clientSet)
			case "DaemonSet":
				return g.waitForDaemonSetSuccess(obj, res.Name)
			case "StatefulSet":
				return g.waitForStatefulSetSuccess(obj, res.Name)
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
	if err != nil {
		logger.Printf("Error in Rss %s", res.Name)
		return err
	} else {
		logger.Printf("Done for %s", res.Name)
		return nil
	}
}

// waitForJob is a helper that waits for a job to complete.
//
// This operates on an event returned from a watcher.
func (g genericPlugin) waitForJob(obj runtime.Object, name string) (bool, error) {
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
func (g genericPlugin) waitForPodSuccess(obj runtime.Object, name string) (bool, error) {
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
func (g genericPlugin) waitForDeploymentSuccess(obj runtime.Object, name string, clientSet kubernetes.Interface) (bool, error) {
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
	newReplicaSet, err := app.GetNewReplicaSet(o, clientSet.AppsV1())
	if err != nil || newReplicaSet == nil {
		return false, err
	}
	expectedReady := *o.Spec.Replicas - app.MaxUnavailable(*o)
	if !(newReplicaSet.Status.ReadyReplicas >= expectedReady) {
		logger.Printf("Deployment is not ready: %s/%s. %d out of %d expected pods are ready", o.Namespace, o.Name, newReplicaSet.Status.ReadyReplicas, expectedReady)
		return false, nil
	}
	return true, nil
}

// waitForDaemonSetSuccess is a helper that waits for a daemonSet to run.
//
// This operates on an event returned from a watcher.
func (g genericPlugin) waitForDaemonSetSuccess(obj runtime.Object, name string) (bool, error) {
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
func (g genericPlugin) waitForStatefulSetSuccess(obj runtime.Object, name string) (bool, error) {
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

// Create generic object in a specific Kubernetes cluster
func (g genericPlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	//Decode the yaml file to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAML(yamlFilePath, unstruct)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode deployment object error")
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	gvk := unstruct.GroupVersionKind()
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}
	if gvk.Kind == "CustomResourceDefinition" {
		//according the helm spec, CRD is created only once, and we raise only warn if we try to do it once more
		resource := helm.KubernetesResource{}
		resource.GVK = gvk
		resource.Name = unstruct.GetName()
		name, err := g.Get(resource, namespace, client)
		if err == nil && name == resource.Name {
			//CRD update is not supported according to Helm spec
			log.Warn(fmt.Sprintf("CRD %s create will be skipped. It already exists", name))
			return name, nil
		}
	}
	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	unstruct.SetLabels(labels)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	plugin.TagPodsIfPresent(unstruct, client.GetInstanceID())

	gvr := mapping.Resource
	var createdObj *unstructured.Unstructured

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		createdObj, err = dynClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), unstruct, metav1.CreateOptions{})
	case meta.RESTScopeNameRoot:
		createdObj, err = dynClient.Resource(gvr).Create(context.TODO(), unstruct, metav1.CreateOptions{})
	default:
		return "", pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + gvk.String())
	}

	if err != nil {
		return "", pkgerrors.Wrap(err, "Create object error")
	}

	return createdObj.GetName(), nil
}

// Update deployment object in a specific Kubernetes cluster
func (g genericPlugin) Update(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	//Decode the yaml file to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAML(yamlFilePath, unstruct)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode deployment object error")
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	gvk := unstruct.GroupVersionKind()
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	if gvk.Kind == "CustomResourceDefinition" {
		resource := helm.KubernetesResource{}
		resource.GVK = gvk
		resource.Name = unstruct.GetName()
		name, err := g.Get(resource, namespace, client)
		if err == nil && name == resource.Name {
			//CRD update is not supported according to Helm spec
			log.Warn(fmt.Sprintf("CRD %s update will be skipped", name))
			return name, nil
		}
	}

	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	unstruct.SetLabels(labels)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	plugin.TagPodsIfPresent(unstruct, client.GetInstanceID())

	gvr := mapping.Resource
	var updatedObj *unstructured.Unstructured

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		updatedObj, err = dynClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), unstruct, metav1.UpdateOptions{})
	case meta.RESTScopeNameRoot:
		updatedObj, err = dynClient.Resource(gvr).Update(context.TODO(), unstruct, metav1.UpdateOptions{})
	default:
		return "", pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + gvk.String())
	}

	if err != nil {
		return "", pkgerrors.Wrap(err, "Update object error")
	}

	return updatedObj.GetName(), nil
}

// Get an existing resource hosted in a specific Kubernetes cluster
func (g genericPlugin) Get(resource helm.KubernetesResource,
	namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: resource.GVK.Group,
		Kind:  resource.GVK.Kind,
	}, resource.GVK.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	gvr := mapping.Resource
	opts := metav1.GetOptions{}
	var unstruct *unstructured.Unstructured
	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		unstruct, err = dynClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), resource.Name, opts)
	case meta.RESTScopeNameRoot:
		unstruct, err = dynClient.Resource(gvr).Get(context.TODO(), resource.Name, opts)
	default:
		return "", pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + resource.GVK.String())
	}

	if err != nil {
		return "", pkgerrors.Wrap(err, "Get object error")
	}

	return unstruct.GetName(), nil
}

// List all existing resources of the GroupVersionKind
// TODO: Implement in seperate patch
func (g genericPlugin) List(gvk schema.GroupVersionKind, namespace string,
	client plugin.KubernetesConnector) ([]helm.KubernetesResource, error) {

	var returnData []helm.KubernetesResource
	return returnData, nil
}

// Delete an existing resource hosted in a specific Kubernetes cluster
func (g genericPlugin) Delete(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) error {
	if namespace == "" {
		namespace = "default"
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: resource.GVK.Group,
		Kind:  resource.GVK.Kind,
	}, resource.GVK.Version)
	if err != nil {
		return pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	gvr := mapping.Resource
	deletePolicy := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if resource.GVK.Kind == "CustomResourceDefinition" {
		//CRD deletion is not supported according to Helm spec
		log.Warn(fmt.Sprintf("CRD %s deletion will be skipped", resource.Name))
		return nil
	}

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		err = dynClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), resource.Name, opts)
	case meta.RESTScopeNameRoot:
		err = dynClient.Resource(gvr).Delete(context.TODO(), resource.Name, opts)
	default:
		return pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + resource.GVK.String())
	}

	if err != nil {
		return pkgerrors.Wrap(err, "Delete object error")
	}
	return nil
}
