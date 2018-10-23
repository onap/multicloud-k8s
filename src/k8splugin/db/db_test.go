// +build unit

/*
Copyright 2018 Intel Corporation.
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

package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestCreateDBClient(t *testing.T) {
	t.Run("Successfully create DB client", func(t *testing.T) {
		expected := &ConsulStore{}

		err := CreateDBClient("consul")
		if err != nil {
			t.Fatalf("CreateDBClient returned an error (%s)", err)
		}
		if reflect.TypeOf(DBconn) != reflect.TypeOf(expected) {
			t.Fatalf("CreateDBClient set DBconn as:\n result=%T\n expected=%T", DBconn, expected)
		}
	})
	t.Run("Fail to create client for unsupported DB", func(t *testing.T) {
		err := CreateDBClient("fakeDB")
		if err == nil {
			t.Fatal("CreateDBClient didn't return an error")
		}
		if !strings.Contains(string(err.Error()), "DB not supported") {
			t.Fatalf("CreateDBClient method returned an error (%s)", err)
		}
	})
}
