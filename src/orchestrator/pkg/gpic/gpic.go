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
	"log"
)

// Cluster has two fields - ClustersWithName and ClustersWithLabel. All the clusters normallly shall fall into either of the two categories.
type Cluster struct {
	ClustersWithName  []ClusterWithName
	ClustersWithLabel []ClusterWithLabel
}

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
func intentResolverHelper(pn, cn, cln string, clustersWithName []ClusterWithName, clustersWithLabel []ClusterWithLabel) ([]ClusterWithName, []ClusterWithLabel) {
	if cln == "" && cn != "" {
		eachClusterWithName := ClusterWithName{pn, cn}
		clustersWithName = append(clustersWithName, eachClusterWithName)
		log.Printf("Added Cluster: %s ", cn)
	}
	if cn == "" && cln != "" {
		eachClusterWithLabel := ClusterWithLabel{pn, cln}
		clustersWithLabel = append(clustersWithLabel, eachClusterWithLabel)
		log.Printf("Added Cluster: %s ", cln)

	}
	return clustersWithName, clustersWithLabel
}

// IntentResolver shall help to resolve the given intent into 2 lists of clusters where the app need to be deployed.
func IntentResolver(intent IntentStruc) Cluster {
	var clustersWithName []ClusterWithName
	var clustersWithLabel []ClusterWithLabel

	for _, eachAllOf := range intent.AllOfArray {
		clustersWithName, clustersWithLabel = intentResolverHelper(eachAllOf.ProviderName, eachAllOf.ClusterName, eachAllOf.ClusterLabelName, clustersWithName, clustersWithLabel)
		if len(eachAllOf.AnyOfArray) > 0 {
			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				clustersWithName, clustersWithLabel = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, clustersWithName, clustersWithLabel)
			}
		}
	}
	if len(intent.AnyOfArray) > 0 {
		for _, eachAnyOf := range intent.AnyOfArray {
			clustersWithName, clustersWithLabel = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, clustersWithName, clustersWithLabel)
		}
	}

	clusters := Cluster{clustersWithName, clustersWithLabel}
	return clusters

}
