// +build unit

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
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	pkgerrors "github.com/pkg/errors"
	"strings"
	"testing"
)

// Implements the mongo.Cursor interface
type mockCursor struct {
	mongo.Cursor
	err   error
	bson  bson.Raw
	count int
}

func (mc *mockCursor) Next(ctx context.Context) bool {
	if mc.count > 0 {
		mc.count = mc.count - 1
		return true
	}
	return false
}

func (mc *mockCursor) DecodeBytes() (bson.Raw, error) {
	return mc.bson, mc.err
}

func (mc *mockCursor) Close(ctx context.Context) error {
	return nil
}

//Implements the functions used currently in mongo.go
type mockCollection struct {
	Err     error
	mCursor mongo.Cursor
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
	opts ...*options.FindOptions) (mongo.Cursor, error) {

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
				"key":  "keyvalue",
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
				"key":  "keyvalue",
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
				"key":  "",
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

			err := m.Create(testCase.input["coll"].(string), testCase.input["key"].(string),
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
		expected      string
	}{
		{
			label: "Successfull Read of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  "keyvalue",
				"tag":  "metadata",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : "b82c4bb1-09ff-6093-4d58-8327b94e1e20",
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
				'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
				'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
				'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
				'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
				'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
				'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
				'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
				'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
				'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
				'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
				'\xfd', '\x00',
			},
			mockColl: &mockCollection{},
			// This is not the document because we are mocking decodeBytes
			expected: "5c115156c9755047e318bbfd",
		},
		{
			label: "UnSuccessfull Read of entry: object not found",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  "keyvalue",
				"tag":  "badtag",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : "b82c4bb1-09ff-6093-4d58-8327b94e1e20",
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
				'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
				'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
				'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
				'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
				'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
				'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
				'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
				'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
				'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
				'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
				'\xfd', '\x00',
			},
			mockColl:      &mockCollection{},
			expectedError: "Error finding objectID",
		},
		{
			label: "UnSuccessfull Read of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  "keyvalue",
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
				"key":  "",
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
			got, err := m.Read(testCase.input["coll"].(string), testCase.input["key"].(string),
				testCase.input["tag"].(string))
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Read method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Read method returned an error (%s)", err)
				}
			} else {
				if bytes.Compare(got, []byte(testCase.expected)) == 0 {
					t.Fatalf("Read returned unexpected data: %s, expected: %s",
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
				"key":  "keyvalue",
				"tag":  "metadata",
			},
			// Binary form of
			// {
			//	"_id" : ObjectId("5c115156777ff85654248ae1"),
			//  "key" : "b82c4bb1-09ff-6093-4d58-8327b94e1e20",
			//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
			// }
			bson: bson.Raw{
				'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
				'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
				'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
				'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
				'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
				'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
				'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
				'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
				'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
				'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
				'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
				'\xfd', '\x00',
			},
			mockColl: &mockCollection{},
		},
		{
			label: "UnSuccessfull Delete of entry",
			input: map[string]interface{}{
				"coll": "collname",
				"key":  "keyvalue",
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
				"key":  "keyvalue",
				"tag":  "tagName",
			},
			bson: bson.Raw{
				'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
				'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
				'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
				'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
				'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
				'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
				'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
				'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
				'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
				'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
				'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
				'\xfd', '\x00',
			},
			mockColl:      &mockCollection{},
			expectedError: "Error finding objectID",
		},
		{
			label: "Missing input fields",
			input: map[string]interface{}{
				"coll": "",
				"key":  "",
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
			err := m.Delete(testCase.input["coll"].(string), testCase.input["key"].(string),
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
		mCursor       *mockCursor
		bson          bson.Raw
		expectedError string
	}{
		{
			label: "Successfully Read all entries",
			input: map[string]interface{}{
				"coll": "collname",
				"tag":  "metadata",
			},
			mockColl: &mockCollection{
				mCursor: &mockCursor{
					// Binary form of
					// {
					//	"_id" : ObjectId("5c115156777ff85654248ae1"),
					//  "key" : "b82c4bb1-09ff-6093-4d58-8327b94e1e20",
					//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
					// }
					bson: bson.Raw{
						'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
						'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
						'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
						'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
						'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
						'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
						'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
						'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
						'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
						'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
						'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
						'\xfd', '\x00',
					},
					count: 1,
				},
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
				mCursor: &mockCursor{
					// Binary form of
					// {
					//	"_id" : ObjectId("5c115156777ff85654248ae1"),
					//  "key" : "b82c4bb1-09ff-6093-4d58-8327b94e1e20",
					//  "metadata" : ObjectId("5c115156c9755047e318bbfd")
					// }
					bson: bson.Raw{
						'\x5a', '\x00', '\x00', '\x00', '\x07', '\x5f', '\x69', '\x64',
						'\x00', '\x5c', '\x11', '\x51', '\x56', '\x77', '\x7f', '\xf8',
						'\x56', '\x54', '\x24', '\x8a', '\xe1', '\x02', '\x6b', '\x65',
						'\x79', '\x00', '\x25', '\x00', '\x00', '\x00', '\x62', '\x38',
						'\x32', '\x63', '\x34', '\x62', '\x62', '\x31', '\x2d', '\x30',
						'\x39', '\x66', '\x66', '\x2d', '\x36', '\x30', '\x39', '\x33',
						'\x2d', '\x34', '\x64', '\x35', '\x38', '\x2d', '\x38', '\x33',
						'\x32', '\x37', '\x62', '\x39', '\x34', '\x65', '\x31', '\x65',
						'\x32', '\x30', '\x00', '\x07', '\x6d', '\x65', '\x74', '\x61',
						'\x64', '\x61', '\x74', '\x61', '\x00', '\x5c', '\x11', '\x51',
						'\x56', '\xc9', '\x75', '\x50', '\x47', '\xe3', '\x18', '\xbb',
						'\xfd', '\x00',
					},
					count: 1,
				},
			},
			expectedError: "Error finding objectID",
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
				return testCase.bson, testCase.mockColl.Err
			}

			_, err := m.ReadAll(testCase.input["coll"].(string), testCase.input["tag"].(string))
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Readall method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Readall method returned an error (%s)", err)
				}
			}
		})
	}
}
