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

package vnfd

import (
	"errors"
	"k8splugin/db"
	"reflect"
	"testing"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockDB struct {
	db.DatabaseConnection
}

func (c *mockDB) CreateEntry(key string, value string) error {
	return nil
}

func (c *mockDB) ReadEntry(key string) (string, bool, error) {

	if key == "vnfd/123e4567-e89b-12d3-a456-426655440000" {
		str := "{\"name\":\"testvnf\"," +
			"\"description\":\"testvnf\"," +
			"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
			"\"service-type\":\"firewall\"}"
		return str, true, nil
	}

	if key == "vnfd/123e4567-e89b-12d3-a456-426655441111" {
		str := "{\"name\":\"testvnf2\"," +
			"\"description\":\"testvnf2\"," +
			"\"uuid\":\"123e4567-e89b-12d3-a456-426655441111\"," +
			"\"service-type\":\"dns\"}"
		return str, true, nil
	}

	return "", false, errors.New("Unable to find Entry")
}

func (c *mockDB) DeleteEntry(key string) error {
	return nil
}

func (c *mockDB) ReadAll(prefix string) ([]string, error) {
	returnVal := []string{prefix + "123e4567-e89b-12d3-a456-426655440000",
		prefix + "123e4567-e89b-12d3-a456-426655441111"}
	return returnVal, nil
}

func TestCreateEntry(t *testing.T) {

	inp := VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "testvnf",
		ServiceType: "firewall",
	}

	db.DBconn = &mockDB{}

	err := inp.createEntry("vnfd/")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {

	vimpl := GetVNFDClient()
	inp := VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "testvnf",
		ServiceType: "firewall",
	}

	db.DBconn = &mockDB{}

	got, err := vimpl.Create(inp)
	if err != nil {
		t.Fatal(err)
	}

	expected := VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655440000",
		Name:        "testvnf",
		Description: "testvnf",
		ServiceType: "firewall",
	}

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("Create VNF returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestList(t *testing.T) {

	vimpl := GetVNFDClient()

	db.DBconn = &mockDB{}

	got, err := vimpl.List()
	if err != nil {
		t.Fatal(err)
	}

	expected := []VNFDefinition{
		{
			UUID:        "123e4567-e89b-12d3-a456-426655440000",
			Name:        "testvnf",
			Description: "testvnf",
			ServiceType: "firewall",
		},
		{
			UUID:        "123e4567-e89b-12d3-a456-426655441111",
			Name:        "testvnf2",
			Description: "testvnf2",
			ServiceType: "dns",
		},
	}

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("List VNFs returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestGet(t *testing.T) {

	vimpl := GetVNFDClient()

	db.DBconn = &mockDB{}

	got, err := vimpl.Get("123e4567-e89b-12d3-a456-426655441111")
	if err != nil {
		t.Fatal(err)
	}

	expected := VNFDefinition{
		UUID:        "123e4567-e89b-12d3-a456-426655441111",
		Name:        "testvnf2",
		Description: "testvnf2",
		ServiceType: "dns",
	}

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("Get VNF returned unexpected body: got %v;"+
			" expected %v", got, expected)
	}
}

func TestDelete(t *testing.T) {

	vimpl := GetVNFDClient()

	db.DBconn = &mockDB{}

	err := vimpl.Delete("123e4567-e89b-12d3-a456-426655441111")
	if err != nil {
		t.Fatal(err)
	}
}
