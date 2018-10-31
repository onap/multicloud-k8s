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

package api

import (
	"encoding/json"
	"k8splugin/vnfd"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var vh vnfdHandler

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockVNFDefinition struct {
	vnfd.VNFDefinitionInterface
}

func (m *mockVNFDefinition) Create(inp vnfd.VNFDefinition) (vnfd.VNFDefinition, error) {
	return vnfd.VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "test description",
		ServiceType: "firewall",
	}, nil
}

func (m *mockVNFDefinition) List() ([]vnfd.VNFDefinition, error) {
	return []vnfd.VNFDefinition{
		{
			UUID:        "123e4567-e89b-12d3-a456-426655440000",
			Name:        "testvnf",
			Description: "test description",
			ServiceType: "firewall",
		},
		{
			UUID:        "123e4567-e89b-12d3-a456-426655441111",
			Name:        "testvnf2",
			Description: "test description",
			ServiceType: "dns",
		}}, nil
}

func (m *mockVNFDefinition) Get(vnfID string) (vnfd.VNFDefinition, error) {
	return vnfd.VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "test description",
		ServiceType: "firewall",
	}, nil
}

func (m *mockVNFDefinition) Delete(vnfID string) error {
	return nil
}

func init() {
	mockClient := mockVNFDefinition{}
	vh = vnfdHandler{vnfdClient: &mockClient}
}

func TestVnfdCreateHandler(t *testing.T) {
	body := `{
		"uuid":"123e4567-e89b-12d3-a456-426655440000",
		"name":"testdomain",
		"description":"test description",
		"service-type":"firewall"
		}`
	reader := strings.NewReader(body)
	req, err := http.NewRequest("POST", "/v1/vnfd", reader)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	hr := http.HandlerFunc(vh.vnfdCreateHandler)

	hr.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected statusCreated return code. Got: %v", rr.Code)
	}

	//Check returned body
	expected := vnfd.VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "test description",
		ServiceType: "firewall",
	}

	got := vnfd.VNFDefinition{}
	json.NewDecoder(rr.Body).Decode(&got)

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("vnfdCreateHandler returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestVnfdListHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "/v1/vnfd", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	hr := http.HandlerFunc(vh.vnfdListHandler)

	hr.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected statusOK return code. Got: %v", rr.Code)
	}

	//Check returned body
	expected := []vnfd.VNFDefinition{
		{
			UUID:        "123e4567-e89b-12d3-a456-426655440000",
			Name:        "testvnf",
			Description: "test description",
			ServiceType: "firewall",
		},
		{
			UUID:        "123e4567-e89b-12d3-a456-426655441111",
			Name:        "testvnf2",
			Description: "test description",
			ServiceType: "dns",
		},
	}

	got := []vnfd.VNFDefinition{}
	json.NewDecoder(rr.Body).Decode(&got)

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("vnfdListHandler returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestVnfdGetHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "/v1/vnfd/123e4567-e89b-12d3-a456-426655440000", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	hr := http.HandlerFunc(vh.vnfdGetHandler)

	hr.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected statusOK return code. Got: %v", rr.Code)
	}

	//Check returned body
	expected := vnfd.VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "test description",
		ServiceType: "firewall",
	}

	got := vnfd.VNFDefinition{}
	json.NewDecoder(rr.Body).Decode(&got)

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("vnfdGetHandler returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestVnfdDeleteHandler(t *testing.T) {

	req, err := http.NewRequest("POST", "/v1/vnfd/123e4567-e89b-12d3-a456-426655440000", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	hr := http.HandlerFunc(vh.vnfdDeleteHandler)

	hr.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected statusNoContent return code. Got: %v", rr.Code)
	}
}
