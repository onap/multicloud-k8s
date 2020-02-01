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

package db

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Implements the functions used currently in mongo.go
type mockCollection struct {
	Err          error
	mCursor      *mongo.Cursor
	mCursorCount int
}

func (c *mockCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {

	if c.Err != nil {
		return nil, c.Err
	}

	return &mongo.InsertOneResult{InsertedID: "_id1234"}, nil
}

func (c *mockCollection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {

	return &mongo.SingleResult{}
}

func (c *mockCollection) FindOneAndUpdate(ctx context.Context, filter interface{},
	update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {

	return &mongo.SingleResult{}
}

func (c *mockCollection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	return nil, c.Err
}

func (c *mockCollection) Find(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {

	return c.mCursor, c.Err
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]interface{}
		mockColl      *mockCollection
		bson          bson.Raw
		expectedError string
	}{
		{
			label: "Successfull creation of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
				"data": "Data In String Format",
			},
			bson:     bson.Raw{'\x08', '\x00', '\x00', '\x00', '\x0A', 'x', '\x00', '\x00'},
			mockColl: &mockCollection{},
		},
		{
			label: "UnSuccessfull creation of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
				"data": "Data In String Format",
			},
			mockColl: &mockCollection{
				Err: pkgerrors.New("DB Error"),
			},
			expectedError: "DB Error",
		},
		{
			label: "Missing input fields",
			input: map[string]interface{}{
				"coll": "",
				"key":  MockKey{Key: ""},
				"tag":  "",
				"data": "",
			},
			expectedError: "No Data to store",
			mockColl:      &mockCollection{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			m, _ := NewMongoStore("name", &mongo.Database{})
			// Override the getCollection function with our mocked version
			getCollection = func(coll string, m *MongoStore) MongoCollection {
				return testCase.mockColl
			}

			decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
				return testCase.bson, testCase.mockColl.Err
			}

			err := m.Create(testCase.input["coll"].(string), testCase.input["key"].(Key),
				testCase.input["tag"].(string), testCase.input["data"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]interface{}
		mockColl      *mockCollection
		bson          bson.Raw
		expectedError string
	}{
		{
			label: "Successfull update of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "metadata",
				"data": "Data In String Format",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
				'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
				'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
				'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
				'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
				'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
				'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
				'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
				'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
				'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
				'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
			},
			mockColl: &mockCollection{},
		},
		{
			label: "Entry does not exist",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
				"data": "Data In String Format",
			},
			mockColl: &mockCollection{
				Err: pkgerrors.New("DB Error"),
			},
			expectedError: "DB Error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			m, _ := NewMongoStore("name", &mongo.Database{})
			// Override the getCollection function with our mocked version
			getCollection = func(coll string, m *MongoStore) MongoCollection {
				return testCase.mockColl
			}

			decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
				return testCase.bson, testCase.mockColl.Err
			}

			err := m.Update(testCase.input["coll"].(string), testCase.input["key"].(Key),
				testCase.input["tag"].(string), testCase.input["data"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRead(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]interface{}
		mockColl      *mockCollection
		bson          bson.Raw
		expectedError string
		expected      []byte
	}{
		{
			label: "Successfull Read of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "metadata",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
				'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
				'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
				'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
				'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
				'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
				'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
				'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
				'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
				'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
				'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
			},
			mockColl: &mockCollection{},
			// This is not the document because we are mocking decodeBytes
			expected: []byte{92, 17, 81, 86, 119, 127, 248, 86, 84, 36, 138, 225},
		},
		{
			label: "UnSuccessfull Read of entry: object not found",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "badtag",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
				'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
				'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
				'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
				'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
				'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
				'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
				'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
				'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
				'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
				'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
			},
			mockColl:      &mockCollection{},
			expectedError: "Error finding objectID",
		},
		{
			label: "UnSuccessfull Read of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
			},
			mockColl: &mockCollection{
				Err: pkgerrors.New("DB Error"),
			},
			expectedError: "DB Error",
		},
		{
			label: "Missing input fields",
			input: map[string]interface{}{
				"coll": "",
				"key":  MockKey{Key: ""},
				"tag":  "",
			},
			expectedError: "Mandatory fields are missing",
			mockColl:      &mockCollection{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			m, _ := NewMongoStore("name", &mongo.Database{})
			// Override the getCollection function with our mocked version
			getCollection = func(coll string, m *MongoStore) MongoCollection {
				return testCase.mockColl
			}

			decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
				return testCase.bson, testCase.mockColl.Err
			}
			got, err := m.Read(testCase.input["coll"].(string), testCase.input["key"].(Key),
				testCase.input["tag"].(string))
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Read method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Read method returned an error (%s)", err)
				}
			} else {
				if bytes.Compare(got, testCase.expected) != 0 {
					t.Fatalf("Read returned unexpected data: %v, expected: %v",
						string(got), testCase.expected)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]interface{}
		mockColl      *mockCollection
		bson          bson.Raw
		expectedError string
	}{
		{
			label: "Successfull Delete of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "metadata",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
				'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
				'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
				'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
				'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
				'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
				'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
				'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
				'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
				'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
				'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
			},
			mockColl: &mockCollection{},
		},
		{
			label: "UnSuccessfull Delete of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
			},
			mockColl: &mockCollection{
				Err: pkgerrors.New("DB Error"),
			},
			expectedError: "DB Error",
		},
		{
			label: "UnSuccessfull Delete, key not found",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  MockKey{Key: "keyvalue"},
				"tag":  "tagName",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
				'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
				'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
				'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
				'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
				'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
				'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
				'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
				'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
				'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
				'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
			},
			mockColl:      &mockCollection{},
			expectedError: "Error finding objectID",
		},
		{
			label: "Missing input fields",
			input: map[string]interface{}{
				"coll": "",
				"key":  MockKey{Key: ""},
				"tag":  "",
			},
			expectedError: "Mandatory fields are missing",
			mockColl:      &mockCollection{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			m, _ := NewMongoStore("name", &mongo.Database{})
			// Override the getCollection function with our mocked version
			getCollection = func(coll string, m *MongoStore) MongoCollection {
				return testCase.mockColl
			}

			decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
				return testCase.bson, testCase.mockColl.Err
			}
			err := m.Delete(testCase.input["coll"].(string), testCase.input["key"].(Key),
				testCase.input["tag"].(string))
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Delete method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestReadAll(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]interface{}
		mockColl      *mockCollection
		bson          bson.Raw
		expectedError string
		expected      map[string][]byte
	}{
		{
			label: "Successfully Read all entries",
			input: map[string]interface{}{
				"coll": "collname",
				"tag":  "metadata",
			},
			mockColl: &mockCollection{
				mCursor: &mongo.Cursor{
					// Binary form of
					// {
					//	"_id" : ObjectId("5c115156777ff85654248ae1"),
					//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
					//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
					// }

					Current: bson.Raw{
						'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
						'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
						'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
						'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
						'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
						'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
						'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
						'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
						'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
						'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
						'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
					},
				},
				mCursorCount: 1,
			},
			expected: map[string][]byte{
				`{"name": "testdef","version": "v1"}`: []byte{
					92, 17, 81, 86, 119, 127, 248, 86, 84, 36, 138, 225},
			},
		},
		{
			label: "UnSuccessfully Read of all entries",
			input: map[string]interface{}{
				"coll": "collname",
				"tag":  "tagName",
			},
			mockColl: &mockCollection{
				Err: pkgerrors.New("DB Error"),
			},
			expectedError: "DB Error",
		},
		{
			label: "UnSuccessfull Readall, tag not found",
			input: map[string]interface{}{
				"coll": "collname",
				"tag":  "tagName",
			},
			mockColl: &mockCollection{
				mCursor: &mongo.Cursor{
					// Binary form of
					// {
					//	"_id" : ObjectId("5c115156777ff85654248ae1"),
					//  "key" : bson.D{{"name","testdef"},{"version","v1"}},
					//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
					// }
					Current: bson.Raw{
						'\x58', '\x00', '\x00', '\x00', '\x03', '\x6b', '\x65', '\x79',
						'\x00', '\x27', '\x00', '\x00', '\x00', '\x02', '\x6e', '\x61',
						'\x6d', '\x65', '\x00', '\x08', '\x00', '\x00', '\x00', '\x74',
						'\x65', '\x73', '\x74', '\x64', '\x65', '\x66', '\x00', '\x02',
						'\x76', '\x65', '\x72', '\x73', '\x69', '\x6f', '\x6e', '\x00',
						'\x03', '\x00', '\x00', '\x00', '\x76', '\x31', '\x00', '\x00',
						'\x07', '\x6d', '\x65', '\x74', '\x61', '\x64', '\x61', '\x74',
						'\x61', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f',
						'\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x07', '\x5f',
						'\x69', '\x64', '\x00', '\x5c', '\x11', '\x51', '\x56', '\x77',
						'\x7f', '\xf8', '\x56', '\x54', '\x24', '\x8a', '\xe1', '\x00',
					},
				},
				mCursorCount: 1,
			},
			expectedError: "Did not find any objects with tag",
		},
		{
			label: "Missing input fields",
			input: map[string]interface{}{
				"coll": "",
				"tag":  "",
			},
			expectedError: "Missing collection or tag name",
			mockColl:      &mockCollection{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			m, _ := NewMongoStore("name", &mongo.Database{})
			// Override the getCollection function with our mocked version
			getCollection = func(coll string, m *MongoStore) MongoCollection {
				return testCase.mockColl
			}

			decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
				return testCase.mockColl.mCursor.Current, testCase.mockColl.Err
			}

			cursorNext = func(ctx context.Context, cursor *mongo.Cursor) bool {
				if testCase.mockColl.mCursorCount > 0 {
					testCase.mockColl.mCursorCount -= 1
					return true
				}
				return false
			}

			cursorClose = func(ctx context.Context, cursor *mongo.Cursor) error {
				return nil
			}

			got, err := m.ReadAll(testCase.input["coll"].(string), testCase.input["tag"].(string))
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Readall method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Readall method returned an error (%s)", err)
				}
			} else {
				if reflect.DeepEqual(got, testCase.expected) == false {
					t.Fatalf("Readall returned unexpected data: %v, expected: %v",
						got, testCase.expected)
				}
			}
		})
	}
}
