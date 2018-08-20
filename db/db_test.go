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
	"testing"
)

func TestCreateDBClient(t *testing.T) {
	oldDBconn := DBconn

	defer func() {
		DBconn = oldDBconn
	}()

	t.Run("Successfully create DB client", func(t *testing.T) {
		expectedDB := ConsulDB{}

		err := CreateDBClient("consul")
		if err != nil {
			t.Fatalf("TestCreateDBClient returned an error (%s)", err)
		}

		if !reflect.DeepEqual(DBconn, &expectedDB) {
			t.Fatalf("TestCreateDBClient set DBconn as:\n result=%v\n expected=%v", DBconn, expectedDB)
		}
	})
}
