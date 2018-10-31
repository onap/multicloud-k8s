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

func TestSerialize(t *testing.T) {

	inp := map[string]interface{}{
		"UUID":        "123e4567-e89b-12d3-a456-426655440000",
		"Data":        "sdaijsdiodalkfjsdlagf",
		"Number":      23,
        "Float":       34.4, 
		"Map": map[string]interface{}{
			"m1": "m1",
			"m2": 2,
            "m3": 3.0,
		},
	}

	got, err := Serialize(inp)
	if err != nil {
		t.Fatal(err)
	}

    expected := "{\"Data\":\"sdaijsdiodalkfjsdlagf\"," +
        "\"Float\":34.4,\"Map\":{\"m1\":\"m1\",\"m2\":2,\"m3\":3}," +
        "\"Number\":23,\"UUID\":\"123e4567-e89b-12d3-a456-426655440000\"}"

	if expected != got {
		t.Errorf("Serialize returned unexpected string: %s;"+
			" expected %sv", got, expected)
	}
}

func TestDeSerialize(t *testing.T) {

    inp := "{\"Data\":\"sdaijsdiodalkfjsdlagf\"," +
        "\"Float\":34.4,\"Map\":{\"m1\":\"m1\",\"m3\":3}," +
        "\"UUID\":\"123e4567-e89b-12d3-a456-426655440000\"}"

    got := make(map[string]interface{})
	err := DeSerialize(inp,&got)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]interface{}{
		"UUID":        "123e4567-e89b-12d3-a456-426655440000",
		"Data":        "sdaijsdiodalkfjsdlagf",
		"Float":       34.4,
		"Map": map[string]interface{}{
			"m1": "m1",
			"m3": 3.0,
		},
	}

	if reflect.DeepEqual(expected, got) == false {
		t.Errorf("Serialize returned unexpected : %s;"+
			" expected %s", got, expected)
	}
}
