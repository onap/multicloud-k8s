/*Copyright © 2020 Intel Corp

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
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var inputFiles []string
var valuesFiles []string

type ResourceContext struct {
	Anchor string `json:"anchor" yaml:"anchor"`
}

type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	UserData1   string `yaml:"userData1,omitempty" json:"userData1,omitempty"`
	UserData2   string `yaml:"userData2,omitempty" json:"userData2,omitempty"`
}

type emcoRes struct {
	Version string                 `yaml:"version" json:"version"`
	Context ResourceContext        `yaml:"resourceContext" json:"resourceContext"`
	Meta    Metadata               `yaml:"metadata" json:"metadata"`
	Spec    map[string]interface{} `yaml:"spec,omitempty" json:"spec,omitempty"`
	File    string                 `yaml:"file,omitempty" json:"file,omitempty"`
	Label   string                 `yaml:"label-name,omitempty" json:"label-name,omitempty"`
}

type emcoBody struct {
	Meta  Metadata               `json:"metadata,omitempty"`
	Label string                 `json:"label-name,omitempty"`
	Spec  map[string]interface{} `json:"spec,omitempty"`
}

type emcoCompositeAppSpec struct {
	Version string `json: "version"`
}

type Resources struct {
	anchor string
	body   []byte
	file   string
}

// RestyClient to use with CLI
type RestyClient struct {
	client *resty.Client
}

var Client RestyClient

// NewRestClient returns a rest client
func NewRestClient() RestyClient {
	// Create a Resty Client
	Client.client = resty.New()
	// Bearer Auth Token for all request
	// Client.client.SetAuthToken()
	// Registering global Error object structure for JSON/XML request
	//Client.client.SetError(&Error{})
	return Client
}

// readResources reads all the resources in the file provided
func readResources() []Resources {
	// TODO: Remove Assumption only one file
	// Open file and Parse to get all resources
	var resources []Resources
	f, err := os.Open(inputFiles[0])
	defer f.Close()
	if err != nil {
		fmt.Printf("Error %s reading file %s\n", err, inputFiles[0])
		return []Resources{}
	}
	if len(valuesFiles) > 0 {
		//Apply template
	}

	dec := yaml.NewDecoder(f)
	// Iterate through all resources in the file
	for {
		var doc emcoRes
		if dec.Decode(&doc) != nil {
			break
		}
		body := &emcoBody{Meta: doc.Meta, Spec: doc.Spec, Label: doc.Label}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return []Resources{}
		}
		var res Resources
		if doc.File != "" {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody, file: doc.File}
		} else {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody}
		}
		resources = append(resources, res)
	}
	return resources
}

//RestClientPost to post to server no multipart
func (r RestyClient) RestClientPost(anchor string, body []byte) error {

	url, err := GetURL(anchor)
	if err != nil {
		return err
	}

	// POST JSON string
	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("URL:", anchor, "Response Code:", resp.StatusCode(), "Response:", resp)
	if resp.StatusCode() >= 201 && resp.StatusCode() <= 299 {
		return nil
	}
	return pkgerrors.Errorf("Server Post Error")
}

//RestClientMultipartPost to post to server with multipart
func (r RestyClient) RestClientMultipartPost(anchor string, body []byte, file string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}

	// Read file for multipart
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error %s reading file %s\n", err, file)
		return err
	}

	// Multipart Post
	formParams := neturl.Values{}
	formParams.Add("metadata", string(body))
	resp, err := r.client.R().
		SetFileReader("file", "filename", bytes.NewReader(f)).
		SetFormDataFromValues(formParams).
		Post(url)

	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("URL:", anchor, "Response Code:", resp.StatusCode(), "Response:", resp)
	if resp.StatusCode() >= 201 && resp.StatusCode() <= 299 {
		return nil
	}
	return pkgerrors.Errorf("Server Multipart Post Error")
}

// RestClientGetAnchor returns get data from anchor
func (r RestyClient) RestClientGetAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	s := strings.Split(anchor, "/")
	if len(s) >= 3 {
		a := s[len(s)-2]
		// Determine if multipart
		if a == "apps" || a == "profiles" || a == "clusters" {
			// Supports only getting metadata
			resp, err := r.client.R().
				SetHeader("Accept", "application/json").
				Get(url)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("URL:", anchor, "Response Code:", resp.StatusCode(), "Response:", resp)
			return nil
		}
	}
	resp, err := r.client.R().
		Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("URL:", anchor, "Response Code:", resp.StatusCode(), "Response:", resp)
	return nil
}

// RestClientGet gets resource
func (r RestyClient) RestClientGet(anchor string, body []byte) error {
	if anchor == "" {
		return pkgerrors.Errorf("Anchor can't be empty")
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" || a == "apply" || a == "approve" || a == "terminate" {
		// No get for these
		return nil
	}
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if e.Meta.Name != "" {
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("mapstruct error")
				return err
			}
			anchor = anchor + "/" + cav.Version
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}

	return r.RestClientGetAnchor(anchor)
}

// RestClientDeleteAnchor returns all resource in the input file
func (r RestyClient) RestClientDeleteAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	resp, err := r.client.R().Delete(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("URL:", anchor, "Response Code:", resp.StatusCode(), "Response:", resp)
	return nil
}

// RestClientDelete calls rest delete command
func (r RestyClient) RestClientDelete(anchor string, body []byte) error {

	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" {
		// Change instantiate to destroy
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "apply" {
		// Change apply to terminate
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "approve" || a == "status" {
		// Approve and status  doesn't have delete
		return nil
	}
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if e.Meta.Name != "" {
		s := strings.Split(anchor, "/")
		a := s[len(s)-1]
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("mapstruct error")
				return err
			}
			anchor = anchor + "/" + cav.Version
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}
	return r.RestClientDeleteAnchor(anchor)
}

// GetURL reads the configuration file to get URL
func GetURL(anchor string) (string, error) {
	var baseUrl string
	s := strings.Split(anchor, "/")
	if len(s) < 1 {
		return "", fmt.Errorf("Invalid Anchor")
	}

	switch s[0] {
	case "cluster-providers":
		if len(s) >= 5 && (s[4] == "networks" || s[4] == "provider-networks" || s[4] == "apply" || s[4] == "terminate") {
			baseUrl = GetNcmURL()
		} else {
			baseUrl = GetClmURL()
		}
	case "controllers":
		baseUrl = GetOrchestratorURL()
	case "projects":
		if len(s) >= 3 && s[2] == "logical-clouds" {
			baseUrl = GetDcmURL()
			break
		}
		if len(s) >= 8 && s[7] == "network-controller-intent" {
			baseUrl = GetOvnactionURL()
			break
		}
		// All other paths go to Orchestrator
		baseUrl = GetOrchestratorURL()
	default:
		return "", fmt.Errorf("Invalid Anchor")
	}
	fmt.Printf(baseUrl)
	return (baseUrl + "/" + anchor), nil
}
