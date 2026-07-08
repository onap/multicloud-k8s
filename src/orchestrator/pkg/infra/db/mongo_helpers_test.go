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

package db

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// helperKey is a sample composite key used by the helper tests below.
type helperKey struct {
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"compositeappversion"`
}

func TestValidateParams(t *testing.T) {
	m := &MongoStore{}

	testCases := []struct {
		label    string
		args     []interface{}
		expected bool
	}{
		{
			label:    "All params valid",
			args:     []interface{}{"coll", "tag", helperKey{Project: "p"}},
			expected: true,
		},
		{
			label:    "Empty string param",
			args:     []interface{}{"coll", ""},
			expected: false,
		},
		{
			label:    "Nil param",
			args:     []interface{}{"coll", nil},
			expected: false,
		},
		{
			label:    "No params",
			args:     []interface{}{},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			got := m.validateParams(testCase.args...)
			if got != testCase.expected {
				t.Fatalf("validateParams returned %v; expected %v", got, testCase.expected)
			}
		})
	}
}

func TestFindFilter(t *testing.T) {
	m := &MongoStore{}

	key := helperKey{Project: "p1", CompositeApp: "ca1", Version: "v1"}
	got, err := m.findFilter(key)
	if err != nil {
		t.Fatalf("findFilter returned an unexpected error: %s", err)
	}

	expected := bson.M{
		"$and": []bson.M{
			{
				"project":             "p1",
				"compositeapp":        "ca1",
				"compositeappversion": "v1",
			},
		},
	}
	if !reflect.DeepEqual(primitive.M(expected), got) {
		t.Fatalf("findFilter returned %v; expected %v", got, expected)
	}
}

func TestFindFilterWithKey(t *testing.T) {
	m := &MongoStore{}

	t.Run("All fields set", func(t *testing.T) {
		key := helperKey{Project: "p1", CompositeApp: "ca1", Version: "v1"}
		got, err := m.findFilterWithKey(key)
		if err != nil {
			t.Fatalf("findFilterWithKey returned an unexpected error: %s", err)
		}
		expected := bson.M{
			"$and": []bson.M{
				{
					"project":             "p1",
					"compositeapp":        "ca1",
					"compositeappversion": "v1",
				},
			},
		}
		if !reflect.DeepEqual(primitive.M(expected), got) {
			t.Fatalf("findFilterWithKey returned %v; expected %v", got, expected)
		}
	})

	t.Run("Empty field replaced with key type", func(t *testing.T) {
		// When a field is empty, findFilterWithKey should replace it with
		// a "key" entry that is the sorted set of key field names.
		key := helperKey{Project: "p1", CompositeApp: "", Version: ""}
		got, err := m.findFilterWithKey(key)
		if err != nil {
			t.Fatalf("findFilterWithKey returned an unexpected error: %s", err)
		}
		expectedKeyField := "{compositeapp,compositeappversion,project,}"
		expected := bson.M{
			"$and": []bson.M{
				{
					"project": "p1",
					"key":     expectedKeyField,
				},
			},
		}
		if !reflect.DeepEqual(primitive.M(expected), got) {
			t.Fatalf("findFilterWithKey returned %v; expected %v", got, expected)
		}
	})
}

func TestUpdateFilter(t *testing.T) {
	m := &MongoStore{}

	key := helperKey{Project: "p1", CompositeApp: "ca1", Version: "v1"}
	got, err := m.updateFilter(key)
	if err != nil {
		t.Fatalf("updateFilter returned an unexpected error: %s", err)
	}
	expected := bson.M{
		"$set": bson.M{
			"project":             "p1",
			"compositeapp":        "ca1",
			"compositeappversion": "v1",
		},
	}
	if !reflect.DeepEqual(primitive.M(expected), got) {
		t.Fatalf("updateFilter returned %v; expected %v", got, expected)
	}
}

func TestCreateKeyField(t *testing.T) {
	m := &MongoStore{}

	key := helperKey{Project: "p1", CompositeApp: "ca1", Version: "v1"}
	got, err := m.createKeyField(key)
	if err != nil {
		t.Fatalf("createKeyField returned an unexpected error: %s", err)
	}
	// Fields are sorted alphabetically by json tag.
	expected := "{compositeapp,compositeappversion,project,}"
	if got != expected {
		t.Fatalf("createKeyField returned %q; expected %q", got, expected)
	}
}

func TestUnmarshal(t *testing.T) {
	m := &MongoStore{}

	t.Run("Successful unmarshal", func(t *testing.T) {
		in := helperKey{Project: "p1", CompositeApp: "ca1", Version: "v1"}
		raw, err := bson.Marshal(in)
		if err != nil {
			t.Fatalf("bson.Marshal setup failed: %s", err)
		}
		var out helperKey
		if err := m.Unmarshal(raw, &out); err != nil {
			t.Fatalf("Unmarshal returned an unexpected error: %s", err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("Unmarshal produced %v; expected %v", out, in)
		}
	})

	t.Run("Invalid bson data", func(t *testing.T) {
		var out helperKey
		err := m.Unmarshal([]byte("not-bson"), &out)
		if err == nil {
			t.Fatalf("Unmarshal expected an error for invalid bson data")
		}
	})
}
