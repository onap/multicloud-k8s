/*
 * Copyright 2020 Intel Corporation, Inc
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

package module

import (
	"encoding/json"
	"fmt"

	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	netutils "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/utils"
	v1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1core "k8s.io/api/core/v1"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	pkgerrors "github.com/pkg/errors"
)

type NfnAnnotation struct {
	CniType   string
	Interface []WorkloadIfIntentSpec
}

const NfnAnnotationKey = "k8s.plugin.opnfv.org/nfn-network"

// ParsePodTemplateNetworkAnnotation parses Pod annotation in PodTemplate
func ParsePodTemplateNetworkAnnotation(pt *v1core.PodTemplateSpec) ([]*nettypes.NetworkSelectionElement, error) {
	netAnnot := pt.Annotations[nettypes.NetworkAttachmentAnnot]
	defaultNamespace := pt.Namespace

	if len(netAnnot) == 0 {
		return nil, pkgerrors.Errorf("No kubernetes network annotation found")
	}

	networks, err := netutils.ParseNetworkAnnotation(netAnnot, defaultNamespace)
	if err != nil {
		return nil, err
	}
	return networks, nil
}

// GetPodTemplateNetworkAnnotation gets Pod Nfn annotation in PodTemplate
func GetPodTemplateNfnAnnotation(pt *v1core.PodTemplateSpec) NfnAnnotation {
	var nfn NfnAnnotation

	nfnAnnot := pt.Annotations[NfnAnnotationKey]
	if len(nfnAnnot) == 0 {
		return nfn
	}

	err := json.Unmarshal([]byte(nfnAnnot), &nfn)
	if err != nil {
		fmt.Println("error unmarshalling nfn annotation\n") //todo - make log warning
	}
	return nfn
}

// GetPodNetworkAnnotation gets Pod Nfn annotation in PodTemplate
func GetPodNfnAnnotation(p *v1core.Pod) NfnAnnotation {
	var nfn NfnAnnotation

	nfnAnnot := p.Annotations[NfnAnnotationKey]
	if len(nfnAnnot) == 0 {
		return nfn
	}

	err := json.Unmarshal([]byte(nfnAnnot), &nfn)
	if err != nil {
		fmt.Println("error unmarshalling nfn annotation\n") //todo - make log warning
	}
	return nfn
}

func addNetworkAnnotation(a nettypes.NetworkSelectionElement, as []*nettypes.NetworkSelectionElement) []*nettypes.NetworkSelectionElement {
	var netElements []*nettypes.NetworkSelectionElement

	found := false
	for _, e := range as {
		if e.Name == a.Name {
			found = true
		}
		netElements = append(netElements, e)
	}
	if !found {
		netElements = append(netElements, &a)
	}

	return netElements
}

// Add the interfaces in the 'new' parameter to the nfn annotation
func newNfnIfs(nfn NfnAnnotation, new []WorkloadIfIntentSpec) NfnAnnotation {
	// Prepare a new interface list - combining the original plus new ones
	var newNfn NfnAnnotation

	if nfn.CniType != CNI_TYPE_OVN4NFV {
		if len(nfn.CniType) > 0 {
			fmt.Printf("Warning - Nfn cnitype was not equal to %v\n", CNI_TYPE_OVN4NFV)
		}
	}
	newNfn.CniType = CNI_TYPE_OVN4NFV

	// update any interfaces already in the list with the updated interface
	for _, i := range nfn.Interface {
		for _, j := range new {
			if len(j.DefaultIf) == 0 {
				j.DefaultIf = "false"
			}
			if i.NetworkName == j.NetworkName && i.IfName == j.IfName {
				i.DefaultIf = j.DefaultIf
				i.IpAddr = j.IpAddr
				i.MacAddr = j.MacAddr
				break
			}
		}
		newNfn.Interface = append(newNfn.Interface, i)
	}

	// add new interfaces not present in original list
	for _, j := range new {
		if len(j.DefaultIf) == 0 {
			j.DefaultIf = "false"
		}
		found := false
		for _, i := range nfn.Interface {
			if i.NetworkName == j.NetworkName && i.IfName == j.IfName {
				found = true
				break
			}
		}
		if !found {
			newNfn.Interface = append(newNfn.Interface, j)
		}
	}
	return newNfn
}

// Add a network annotation to the resource
func AddNetworkAnnotation(r interface{}, a nettypes.NetworkSelectionElement) error {

	//fmt.Printf("NETWORK ANNOTATION: %v\n", a)
	switch o := r.(type) {
	case *batch.Job:
		fmt.Printf("GOT A JOB\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil
	case *batchv1beta1.CronJob:
		fmt.Printf("GOT A CRONJOB\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.JobTemplate.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.JobTemplate.Spec.Template.Annotations == nil {
			o.Spec.JobTemplate.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.JobTemplate.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil
	case *v1.DaemonSet:
		fmt.Printf("GOT A DAEMON SET\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil
	case *v1.Deployment:
		fmt.Printf("GOT A DEPLOYMENT\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil
	case *v1.ReplicaSet:
		fmt.Printf("GOT A REPLICASET\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil
	case *v1.StatefulSet:
		fmt.Printf("GOT A STATEFULSET\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)

		return nil
	case *v1core.Pod:
		fmt.Printf("GOT A POD\n")
		netAnnotation, _ := netutils.ParsePodNetworkAnnotation(o)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Annotations == nil {
			o.Annotations = make(map[string]string)
		}
		o.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)

		return nil
	case *v1core.ReplicationController:
		fmt.Printf("GOT A REPLICATION CONTROLLER\n")
		netAnnotation, _ := ParsePodTemplateNetworkAnnotation(o.Spec.Template)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return nil

	default:
		fmt.Printf("GOT AN UNKNOWN\n")
	}

	return pkgerrors.Errorf("Unsupported resource")
}

// Add an nfn annotation to the resource
func AddNfnAnnotation(r interface{}, new []WorkloadIfIntentSpec) error {

	switch o := r.(type) {
	case *batch.Job:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *batchv1beta1.CronJob:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.JobTemplate.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.JobTemplate.Spec.Template.Annotations == nil {
			o.Spec.JobTemplate.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.JobTemplate.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1.DaemonSet:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1.Deployment:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1.ReplicaSet:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1.StatefulSet:
		nfnAnnotation := GetPodTemplateNfnAnnotation(&o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1core.Pod:
		nfnAnnotation := GetPodNfnAnnotation(o)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Annotations == nil {
			o.Annotations = make(map[string]string)
		}
		o.Annotations[NfnAnnotationKey] = string(j)
		return nil
	case *v1core.ReplicationController:
		nfnAnnotation := GetPodTemplateNfnAnnotation(o.Spec.Template)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err) // log an info or warning
			break
		}
		if o.Spec.Template.Annotations == nil {
			o.Spec.Template.Annotations = make(map[string]string)
		}
		o.Spec.Template.Annotations[NfnAnnotationKey] = string(j)
		return nil

	default:
		fmt.Printf("GOT AN UNKNOWN\n")
	}

	return pkgerrors.Errorf("Unsupported resource")
}
