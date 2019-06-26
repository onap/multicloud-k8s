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

package rb

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"os"
	"path/filepath"

	"encoding/base64"

	pkgerrors "github.com/pkg/errors"
	"log"
)

// ConfigTemplate contains the parameters needed for ConfigTemplates
type ConfigTemplate struct {
	TemplateName string `json:"template-name"`
	Description  string `json:"description"`
	ChartName    string
}

// ConfigTemplateManager is an interface exposes the resource bundle  ConfigTemplate functionality
type ConfigTemplateManager interface {
	Create(rbName, rbVersion string, p ConfigTemplate) error
	Get(rbName, rbVersion, templateName string) (ConfigTemplate, error)
	Delete(rbName, rbVersion, templateName string) error
	Upload(rbName, rbVersion, templateName string, inp []byte) error
}

// ConfigTemplateKey is key struct
type ConfigTemplateKey struct {
	RBName       string `json:"rb-name"`
	RBVersion    string `json:"rb-version"`
	TemplateName string `json:"template-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk ConfigTemplateKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ConfigTemplateClient implements the  ConfigTemplateManager
// It will also be used to maintain some localized state
type ConfigTemplateClient struct {
	storeName  string
	tagMeta    string
	tagContent string
}

// NewConfigTemplateClient returns an instance of the  ConfigTemplateClient
// which implements the  ConfigTemplateManager
func NewConfigTemplateClient() *ConfigTemplateClient {
	return &ConfigTemplateClient{
		storeName:  "rbdef",
		tagMeta:    "metadata",
		tagContent: "content",
	}
}

// Create an entry for the resource bundle  ConfigTemplate in the database
func (v *ConfigTemplateClient) Create(rbName, rbVersion string, p ConfigTemplate) error {

	log.Printf("[ConfigiTemplate]: create %s", rbName)
	// Name is required
	if p.TemplateName == "" {
		return pkgerrors.New("Name is required for Resource Bundle  ConfigTemplate")
	}

	//Check if  ConfigTemplate already exists
	_, err := v.Get(rbName, rbVersion, p.TemplateName)
	if err == nil {
		return pkgerrors.New(" ConfigTemplate already exists for this Definition")
	}

	//Check if provided resource bundle information is valid
	_, err = NewDefinitionClient().Get(rbName, rbVersion)
	if err != nil {
		return pkgerrors.Errorf("Invalid Resource Bundle ID provided: %s", err.Error())
	}

	key := ConfigTemplateKey{
		RBName:       rbName,
		RBVersion:    rbVersion,
		TemplateName: p.TemplateName,
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, p)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating  ConfigTemplate DB Entry")
	}

	return nil
}

// Get returns the Resource Bundle  ConfigTemplate for corresponding ID
func (v *ConfigTemplateClient) Get(rbName, rbVersion, templateName string) (ConfigTemplate, error) {
	key := ConfigTemplateKey{
		RBName:       rbName,
		RBVersion:    rbVersion,
		TemplateName: templateName,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
	if err != nil {
		return ConfigTemplate{}, pkgerrors.Wrap(err, "Get ConfigTemplate")
	}

	//value is a byte array
	if value != nil {
		template := ConfigTemplate{}
		err = db.DBconn.Unmarshal(value, &template)
		if err != nil {
			return ConfigTemplate{}, pkgerrors.Wrap(err, "Unmarshaling  ConfigTemplate Value")
		}
		return template, nil
	}

	return ConfigTemplate{}, pkgerrors.New("Error getting ConfigTemplate")
}

// Delete the Resource Bundle  ConfigTemplate from database
func (v *ConfigTemplateClient) Delete(rbName, rbVersion, templateName string) error {
	key := ConfigTemplateKey{
		RBName:       rbName,
		RBVersion:    rbVersion,
		TemplateName: templateName,
	}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ConfigTemplate")
	}

	err = db.DBconn.Delete(v.storeName, key, v.tagContent)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete  ConfigTemplate Content")
	}

	return nil
}

// Upload the contents of resource bundle into database
func (v *ConfigTemplateClient) Upload(rbName, rbVersion, templateName string, inp []byte) error {

	log.Printf("[ConfigTemplate]: Upload %s", templateName)
	key := ConfigTemplateKey{
		RBName:       rbName,
		RBVersion:    rbVersion,
		TemplateName: templateName,
	}
	//ignore the returned data here.
	t, err := v.Get(rbName, rbVersion, templateName)
	if err != nil {
		return pkgerrors.Errorf("Invalid  ConfigTemplate Name  provided %s", err.Error())
	}

	err = isTarGz(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Errorf("Error in file format %s", err.Error())
	}

	chartBasePath, err := ExtractTarBall(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Wrap(err, "Extracting Template")
	}

	finfo, err := ioutil.ReadDir(chartBasePath)
	if err != nil {
		return pkgerrors.Wrap(err, "Detecting chart name")
	}

	//Store the first directory with Chart.yaml found as the chart name
	for _, f := range finfo {
		if f.IsDir() {
			//Check if Chart.yaml exists
			if _, err = os.Stat(filepath.Join(chartBasePath, f.Name(), "Chart.yaml")); err == nil {
				t.ChartName = f.Name()
				break
			}
		}
	}
	if t.ChartName == "" {
		return pkgerrors.New("Invalid template no Chart.yaml file found")
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, t)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating  ConfigTemplate DB Entry")
	}

	//Encode given byte stream to text for storage
	encodedStr := base64.StdEncoding.EncodeToString(inp)
	err = db.DBconn.Create(v.storeName, key, v.tagContent, encodedStr)
	if err != nil {
		return pkgerrors.Errorf("Error uploading data to db %s", err.Error())
	}

	return nil
}

// Download the contents of the ConfigTemplate from DB
// Returns a byte array of the contents
func (v *ConfigTemplateClient) Download(rbName, rbVersion, templateName string) ([]byte, error) {

	log.Printf("[ConfigTemplate]: Download %s", templateName)
	//ignore the returned data here
	//Check if rb is valid
	_, err := v.Get(rbName, rbVersion, templateName)
	if err != nil {
		return nil, pkgerrors.Errorf("Invalid  ConfigTemplate Name provided: %s", err.Error())
	}

	key := ConfigTemplateKey{
		RBName:       rbName,
		RBVersion:    rbVersion,
		TemplateName: templateName,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagContent)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Resource ConfigTemplate content")
	}

	if value != nil {
		//Decode the string from base64
		out, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Decode base64 string")
		}

		if out != nil && len(out) != 0 {
			return out, nil
		}
	}
	return nil, pkgerrors.New("Error downloading  ConfigTemplate content")
}
