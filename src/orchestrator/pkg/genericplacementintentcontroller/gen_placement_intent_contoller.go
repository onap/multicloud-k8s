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

package genericplacementintentcontroller

import (
	"log"
)

type Cluster struct {
	ClustersWithName  []ClusterWithName
	ClustersWithLabel []ClusterWithLabel
}

type ClusterWithName struct {
	ProviderName string
	ClusterName  string
}

type ClusterWithLabel struct {
	ProviderName string
	ClusterLabel string
}

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

func IntentResolver(appName string, AllOfArray []AllOf, AnyOfArray []AnyOf) Cluster {
	var clustersWithLabel []ClusterWithLabel
	var clustersWithName []ClusterWithName

	for _, eachAllOf := range AllOfArray {
		if eachAllOf.ClusterLabelName == "" && eachAllOf.ClusterName != "" {
			eachClusterWithName := ClusterWithName{eachAllOf.ProviderName, eachAllOf.ClusterName}
			clustersWithName = append(clustersWithName, eachClusterWithName)
			log.Printf("Added Cluster: %s ", eachAllOf.ClusterName)
		}
		if eachAllOf.ClusterName == "" && eachAllOf.ClusterLabelName != "" {
			eachClusterWithLabel := ClusterWithLabel{eachAllOf.ProviderName, eachAllOf.ClusterLabelName}
			clustersWithLabel = append(clustersWithLabel, eachClusterWithLabel)
			log.Printf("Added Cluster: %s ", eachAllOf.ClusterLabelName)

		}
		if len(eachAllOf.AnyOfArray) > 0 {
			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				if eachAnyOf.ClusterLabelName == "" && eachAnyOf.ClusterName != "" {
					eachClusterWithName := ClusterWithName{eachAnyOf.ProviderName, eachAnyOf.ClusterName}
					clustersWithName = append(clustersWithName, eachClusterWithName)
					log.Printf("Added Cluster: %s ", eachAnyOf.ClusterName)
				}
				if eachAnyOf.ClusterName == "" && eachAnyOf.ClusterLabelName != "" {
					eachClusterWithLabel := ClusterWithLabel{eachAnyOf.ProviderName, eachAnyOf.ClusterLabelName}
					clustersWithLabel = append(clustersWithLabel, eachClusterWithLabel)
					log.Printf("Added Cluster: %s ", eachAnyOf.ClusterLabelName)
				}
			}

		}

	}

	if len(AnyOfArray) > 0 {
		for _, eachAnyOf := range AnyOfArray {
			if eachAnyOf.ClusterLabelName == "" && eachAnyOf.ClusterName != "" {
				eachClusterWithName := ClusterWithName{eachAnyOf.ProviderName, eachAnyOf.ClusterName}
				clustersWithName = append(clustersWithName, eachClusterWithName)
				log.Printf("Added Cluster: %s ", eachAnyOf.ClusterName)
			}
			if eachAnyOf.ClusterName == "" && eachAnyOf.ClusterLabelName != "" {
				eachClusterWithLabel := ClusterWithLabel{eachAnyOf.ProviderName, eachAnyOf.ClusterLabelName}
				clustersWithLabel = append(clustersWithLabel, eachClusterWithLabel)
				log.Printf("Added Cluster: %s ", eachAnyOf.ClusterLabelName)
			}
		}
	}

	clusters := Cluster{clustersWithName, clustersWithLabel}
	return clusters

}
