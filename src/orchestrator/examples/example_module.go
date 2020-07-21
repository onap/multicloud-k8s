/*
 * Copyright 2018 Intel Corporation, Inc
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

package test

import (
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
	"log"
)

// ExampleClient_Project to test Project
func ExampleClient_Project() {
	// Get handle to the client
	c := moduleLib.NewClient()
	// Check if project is initialized
	if c.Project == nil {
		log.Println("Project is Uninitialized")
		return
	}
	// Perform operations on Project Module
	// POST request (exists == false)
	_, err := c.Project.CreateProject(moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, false)
	if err != nil {
		log.Println(err)
		return
	}
	// PUT request (exists == true)
	_, err = c.Project.CreateProject(moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, true)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = c.Project.GetProject("test")
	if err != nil {
		log.Println(err)
		return
	}
	err = c.Project.DeleteProject("test")
	if err != nil {
		log.Println(err)
	}
}
