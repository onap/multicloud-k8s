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
	"github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	pkgerrors "github.com/pkg/errors"
	"log"
)

// Clusters has 1 field - a list of ClusterNames
type Clusters struct {
	ClustersWithName []ClusterWithName
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
func intentResolverHelper(pn, cn, cln string, clustersWithName []ClusterWithName) ([]ClusterWithName, error) {
	if cln == "" && cn != "" {
		eachClusterWithName := ClusterWithName{pn, cn}
		clustersWithName = append(clustersWithName, eachClusterWithName)
		log.Printf("Added Cluster: %s ", cn)
	}
	if cn == "" && cln != "" {
		//Finding cluster names for the clusterlabel
		clusterNamesList, err := cluster.NewClusterClient().GetClustersWithLabel(pn, cln)
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
func IntentResolver(intent IntentStruc) (Clusters, error) {
	var clustersWithName []ClusterWithName
	var err error

	for _, eachAllOf := range intent.AllOfArray {
		clustersWithName, err = intentResolverHelper(eachAllOf.ProviderName, eachAllOf.ClusterName, eachAllOf.ClusterLabelName, clustersWithName)
		if err != nil {
			return Clusters{}, pkgerrors.Wrap(err, "intentResolverHelper error")
		}
		if len(eachAllOf.AnyOfArray) > 0 {
			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				clustersWithName, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, clustersWithName)
				if err != nil {
					return Clusters{}, pkgerrors.Wrap(err, "intentResolverHelper error")
				}
			}
		}
	}
	if len(intent.AnyOfArray) > 0 {
		for _, eachAnyOf := range intent.AnyOfArray {
			clustersWithName, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, clustersWithName)
			if err != nil {
				return Clusters{}, pkgerrors.Wrap(err, "intentResolverHelper error")
			}
		}
	}
	clusters := Clusters{clustersWithName}
	return clusters, nil
}
