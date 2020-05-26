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

package gpic

/*
 gpic stands for GenericPlacementIntent Controller.
 This file pertains to the implementation and handling of generic placement intents
*/

import (
	ncmmodule "github.com/onap/multicloud-k8s/src/ncm/pkg/module"
	pkgerrors "github.com/pkg/errors"
	"log"
)

// Clusters has 1 field - a list of ClusterNames
// TODO: This shall contain the mandatory cluster and group cluster.
// Introduce two additional types - mandatoryClusterList and list of groups

type ClusterGroup struct {
	OptionalClusters []ClusterWithName
	GroupName string
}

type ClusterList struct {
	MandatoryClusters []ClusterWithName
	ClusterGroups []ClusterGroup
}
// type Clusters struct {
// 	ClustersWithName []ClusterWithName
// }

// ClusterWithName has two fields - ProviderName and ClusterName
type ClusterWithName struct {
	ProviderName string
	ClusterName  string
}

// ClusterWithLabel has two fields - ProviderName and ClusterLabel
type ClusterWithLabel struct {
	ProviderName string
	ClusterLabel string
}

// IntentStruc consists of AllOfArray and AnyOfArray
type IntentStruc struct {
	AllOfArray []AllOf `json:"allOf,omitempty"`
	AnyOfArray []AnyOf `json:"anyOf,omitempty"`
}

// AllOf consists if ProviderName, ClusterName, ClusterLabelName and AnyOfArray. Any of them can be empty
type AllOf struct {
	ProviderName     string  `json:"provider-name,omitempty"`
	ClusterName      string  `json:"cluster-name,omitempty"`
	ClusterLabelName string  `json:"cluster-label-name,omitempty"`
	AnyOfArray       []AnyOf `json:"anyOf,omitempty"`
}

// AnyOf consists of Array of ProviderName & ClusterLabelNames
type AnyOf struct {
	ProviderName     string `json:"provider-name,omitempty"`
	ClusterName      string `json:"cluster-name,omitempty"`
	ClusterLabelName string `json:"cluster-label-name,omitempty"`
}

// intentResolverHelper helps to populate the cluster lists
func intentResolverHelper(pn, cn, cln string, clustersWithName []ClusterWithName) ([]ClusterWithName, error) {
	if cln == "" && cn != "" {
		eachClusterWithName := ClusterWithName{pn, cn}
		clustersWithName = append(clustersWithName, eachClusterWithName)
		log.Printf("Added Cluster: %s ", cn)
	}
	if cn == "" && cln != "" {
		//Finding cluster names for the clusterlabel
		clusterNamesList, err := ncmmodule.NewClusterClient().GetClustersWithLabel(pn, cln)
		if err != nil {
			return []ClusterWithName{}, pkgerrors.Wrap(err, "Error getting clusterLabels")
		}
		// Populate the clustersWithName array with the clusternames found above
		for _, eachClusterName := range clusterNamesList {
			eachClusterWithPN := ClusterWithName{pn, eachClusterName}
			clustersWithName = append(clustersWithName, eachClusterWithPN)
			log.Printf("Added Cluster :: %s through its label: %s ", eachClusterName, cln)
		}
	}
	return clustersWithName, nil
}

// IntentResolver shall help to resolve the given intent into 2 lists of clusters where the app need to be deployed.
// TODO: This is where the real resolution takes place, the o/p shall now be a composite data type of list of mandatory clusters and group clusters. Each group cluster shall be a string of cluster names and there should be an array of group cluster. The AppContext shall have two different methods to add mandatory cluster and add group clusters.
func IntentResolver(intent IntentStruc) (ClusterList, error) {
	var mc []ClusterWithName
	var err error
	var cg []ClusterGroup

	for _, eachAllOf := range intent.AllOfArray {
		mc, err = intentResolverHelper(eachAllOf.ProviderName, eachAllOf.ClusterName, eachAllOf.ClusterLabelName, mc)
		if err != nil {
			return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
		}
		if len(eachAllOf.AnyOfArray) > 0 {

			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				var opc []ClusterWithName
				opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
				if err != nil {
					return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
				}
				eachClustergroup := ClusterGroup{OptionalClusters:opc, GroupName:eachAnyOf.ClusterLabelName}
				cg = append(cg, eachClustergroup)
			}
		}
	}
	if len(intent.AnyOfArray) > 0 {
		var opc []ClusterWithName
		for _, eachAnyOf := range intent.AnyOfArray {
			opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
			if err != nil {
				return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
			}
			eachClustergroup := ClusterGroup{OptionalClusters:opc, GroupName:eachAnyOf.ClusterLabelName}
			cg = append(cg, eachClustergroup)
		}
	}
	clusterList := ClusterList{MandatoryClusters:mc, ClusterGroups:cg}
	return clusterList, nil
}
